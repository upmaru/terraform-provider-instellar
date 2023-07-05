resource "instellar_cluster" "main" {
  name           = "%s"
  provider_name  = "aws"
  region         = "ap-southeast-1"
  endpoint       = "127.0.0.1:8443"
  password_token = "some-password-or-token"
}

resource "instellar_component" "postgres_db" {
  name           = "some-db"
  provider_name  = "aws"
  driver         = "database/postgresql"
  driver_version = "15.2"
  cluster_ids = [
    instellar_cluster.main.id
  ]
  channels = ["develop", "main"]
  credential {
    username = "postgres"
    password = "postgres"
    database = "postgres"
    host     = "localhost"
    port     = 5432
  }
}