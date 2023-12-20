# Terraform Provider Instellar (Terraform Plugin Framework)

Terraform Provider for instellar module. Instellar is a component of [Opsmaru](https://opsmaru.com)

## Development

Create a `.envrc` file with the following:

```bash
export INSTELLAR_HOST="http://localhost:4000"
export INSTELLAR_AUTH_TOKEN=""
```

Make sure you issue a credential from your local instellar instance and fill `INSTELLAR_AUTH_TOKEN` with it.

```shell
make testacc
```
