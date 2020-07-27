# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.13.4-stretch as builder
WORKDIR /go/src/github.com/kubernetes-sigs/aws-efs-csi-driver

# Cache go modules
ENV GOPROXY=direct
COPY go.mod .
COPY go.sum .
RUN go mod download

ADD . .
ARG EFS_CLIENT_SOURCE
ENV EFS_CLIENT_SOURCE=$EFS_CLIENT_SOURCE
RUN make aws-efs-csi-driver

FROM amazonlinux:2.0.20200602.0
RUN yum install util-linux-2.30.2-2.amzn2.0.4.x86_64 amazon-efs-utils-1.26-3.amzn2.noarch -y

# At image build time, static files installed by efs-utils in the config directory, i.e. CAs file, need
# to be saved in another place so that the other stateful files created at runtime, i.e. private key for
# client certificate, in the same config directory can be persisted to host with a host path volume.
# Otherwise creating a host path volume for that directory will clean up everything inside at the first time.
# Those static files need to be copied back to the config directory when the driver starts up.
RUN cp -r /etc/amazon/efs /etc/amazon/efs-static-files

COPY --from=builder /go/src/github.com/kubernetes-sigs/aws-efs-csi-driver/bin/aws-efs-csi-driver /bin/aws-efs-csi-driver
COPY THIRD-PARTY /

ENTRYPOINT ["/bin/aws-efs-csi-driver"]
