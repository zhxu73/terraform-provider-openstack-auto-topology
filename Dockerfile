# syntax=docker/dockerfile:1
FROM golang:1.16 as build
WORKDIR /terraform-provider-openstack-auto-topology
COPY ./openstack/ /terraform-provider-openstack-auto-topology/openstack/
COPY ./provider/ /terraform-provider-openstack-auto-topology/provider/
COPY ./go.mod ./go.sum ./main.go /terraform-provider-openstack-auto-topology/
RUN go build -o terraform-provider-openstack-auto-topology

FROM scratch
# this is a data image that just include the provider binary, this is useful when you want to include the provider binary
# in other container images.
COPY --from=build /terraform-provider-openstack-auto-topology/terraform-provider-openstack-auto-topology /terraform-provider-openstack-auto-topology
