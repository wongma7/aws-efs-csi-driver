package e2e

import (
	"fmt"
	"sync"

	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/storage/testpatterns"
	"k8s.io/kubernetes/test/e2e/storage/testsuites"
)

var (
	ClusterName  string
	Region       string
	fileSystemId string
	deployed     bool
)

type efsDriver struct {
	driverInfo testsuites.DriverInfo
}

var _ testsuites.TestDriver = &efsDriver{}

// TODO implement Inline (unless it's redundant) and DynamicPV
// var _ testsuites.InlineVolumeTestDriver = &efsDriver{}
var _ testsuites.PreprovisionedPVTestDriver = &efsDriver{}

func InitEFSCSIDriver() testsuites.TestDriver {
	return &efsDriver{
		driverInfo: testsuites.DriverInfo{
			Name:                 "efs.csi.aws.com",
			SupportedFsType:      sets.NewString(""),
			SupportedMountOption: sets.NewString("tls", "ro"),
			Capabilities: map[testsuites.Capability]bool{
				testsuites.CapPersistence: true,
				testsuites.CapExec:        true,
				testsuites.CapMultiPODs:   true,
				testsuites.CapRWX:         true,
			},
		},
	}
}

func (e *efsDriver) GetDriverInfo() *testsuites.DriverInfo {
	return &e.driverInfo
}

func (e *efsDriver) SkipUnsupportedTest(testpatterns.TestPattern) {}

func (e *efsDriver) PrepareTest(f *framework.Framework) (*testsuites.PerTestConfig, func()) {
	cancelPodLogs := testsuites.StartPodLogs(f)

	return &testsuites.PerTestConfig{
			Driver:    e,
			Prefix:    "efs",
			Framework: f,
		}, func() {
			cancelPodLogs()
		}
}

func (e *efsDriver) CreateVolume(config *testsuites.PerTestConfig, volType testpatterns.TestVolType) testsuites.TestVolume {
	return nil
}

func (e *efsDriver) GetPersistentVolumeSource(readOnly bool, fsType string, volume testsuites.TestVolume) (*v1.PersistentVolumeSource, *v1.VolumeNodeAffinity) {
	pvSource := v1.PersistentVolumeSource{
		CSI: &v1.CSIPersistentVolumeSource{
			Driver:       e.driverInfo.Name,
			VolumeHandle: fileSystemId,
		},
	}
	return &pvSource, nil
}

// List of testSuites to be executed in below loop
var csiTestSuites = []func() testsuites.TestSuite{
	testsuites.InitVolumesTestSuite,
	testsuites.InitVolumeIOTestSuite,
	testsuites.InitVolumeModeTestSuite,
	testsuites.InitSubPathTestSuite,
	testsuites.InitProvisioningTestSuite,
	testsuites.InitMultiVolumeTestSuite,
}

var _ = ginkgo.SynchronizedAfterSuite(func() {
	// allNodesBody: do nothing because only node 1 needs to delete the EFS
}, func() {
	// Delete the EFS filesystem *once* for all tests (It blocks)
	// Do nothing if it was never created if none of these EFS tests were part of
	// the test suite's focus
	if fileSystemId != "" {
		ginkgo.By(fmt.Sprintf("Deleting EFS filesystem %q", fileSystemId))
		c := NewCloud(Region)
		err := c.DeleteFileSystem(fileSystemId)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("failed to delete EFS filesystem: %s", err))
		}
	}

	if deployed {
		ginkgo.By("Cleaning up EFS CSI driver")
		framework.RunKubectlOrDie("delete", "-k", "github.com/kubernetes-sigs/aws-efs-csi-driver/deploy/kubernetes/overlays/stable/?ref=master")
	}
})

var _ = ginkgo.Describe("[efs-csi] EFS CSI", func() {
	var efsWg sync.WaitGroup
	efsWg.Add(1)
	var efsOnce sync.Once
	ginkgo.BeforeEach(func() {
		// Create the EFS filesystem *once* for all tests (It blocks)
		// The reason creation should be BeforeEach but deletion AfterSuite is
		// because if creation were BeforeSuite it would be unnecessarily attempted
		// even if none of these EFS tests were part of the test suite's focus
		efsOnce.Do(func() {
			defer efsWg.Done()
			ginkgo.By(fmt.Sprintf("Creating EFS filesystem in region %q for cluster %q", Region, ClusterName))
			if Region == "" || ClusterName == "" {
				ginkgo.Fail("failed to create EFS filesystem: both Region and ClusterName must be non-empty")
			}
			c := NewCloud(Region)
			id, err := c.CreateFileSystem(ClusterName)
			if err != nil {
				ginkgo.Fail(fmt.Sprintf("failed to create EFS filesystem: %s", err))
			}
			fileSystemId = id
			ginkgo.By(fmt.Sprintf("Created EFS filesystem %q in region %q for cluster %q", fileSystemId, Region, ClusterName))
		})
		efsWg.Wait()
		if fileSystemId == "" {
			ginkgo.Fail("EFS creation failed")
		}
	})

	var driverWg sync.WaitGroup
	driverWg.Add(1)
	var driverOnce sync.Once
	ginkgo.BeforeEach(func() {
		// Deploy the EFS CSI driver *once* for all tests (It blocks)
		// Alternatively, could deploy the driver for each test in PrepareTest
		driverOnce.Do(func() {
			defer driverWg.Done()
			cs, err := framework.LoadClientset()
			framework.ExpectNoError(err, "loading kubernetes clientset")
			_, err = cs.StorageV1beta1().CSIDrivers().Get("efs.csi.aws.com", metav1.GetOptions{})
			if err == nil {
				// CSIDriver exists, assume driver has already been deployed
				ginkgo.By("Using already-deployed EFS CSI driver")
			} else if err != nil && !apierrors.IsNotFound(err) {
				// Non-NotFound errors are unexpected
				framework.ExpectNoError(err, "getting csidriver efs.csi.aws.com")
			} else {
				ginkgo.By("Deploying EFS CSI driver")
				framework.RunKubectlOrDie("apply", "-k", "github.com/kubernetes-sigs/aws-efs-csi-driver/deploy/kubernetes/overlays/stable/?ref=master")
			}
			deployed = true
		})
		driverWg.Wait()
		if !deployed {
			ginkgo.Fail("driver deployment failed")
		}
	})

	driver := InitEFSCSIDriver()
	ginkgo.Context(testsuites.GetDriverNameWithFeatureTags(driver), func() {
		testsuites.DefineTestSuite(driver, csiTestSuites)
	})
})
