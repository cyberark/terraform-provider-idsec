# Example: Read CCE GCP identity federation parameters
# This example demonstrates how to retrieve workload identity federation
# parameters for active services from CCE GCP

# Read identity params
# This data source retrieves the tenant ID and identity federation parameters
# for each active service (e.g., dpa, sca)
# No input parameters are required
data "idsec_cce_gcp_identity_params" "identity_federation_example" {

}
