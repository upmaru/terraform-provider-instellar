resource "instellar_cluster" "example" {
  name            = "some-unique-name"
  provider_name   = "aws"
  region          = "ap-southeast-1"
  endpoint        = "127.0.0.1:8443"
  passsword_token = "join-token-from-cluster-setup"
}
