resource "instellar_storage" "this" {
  host              = "s3.amazonaws.com"
  bucket            = "mybucket"
  region            = "ap-southeast-1"
  access_key_id     = "somekey"
  secret_access_key = "somesecret"
}