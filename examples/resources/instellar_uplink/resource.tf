resource "instellar_cluster" "main" {
  name           = "some-unique-cluster-name"
  provider_name  = "aws"
  region         = "ap-southeast-1"
  endpoint       = "127.0.0.1:8443"
  password_token = "some-password-or-token"
}

resource "instellar_uplink" "this" {
  channel_slug = "master"
  cluster_id   = instellar_cluster.main.id
  database_url = "postgresql://user:pass@localhost:5432/some_db_example"
}