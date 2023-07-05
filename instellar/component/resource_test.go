package component_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-instellar/internal/acceptance"
)

func TestAccComponentResource(t *testing.T) {
	clusterUUID := uuid.New()
	clusterNameSegments := strings.Split(clusterUUID.String(), "-")
	clusterNameSlug := clusterNameSegments[0]

	componentName := fmt.Sprintf("%s-db", clusterNameSlug)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildConfig(clusterNameSlug, componentName, `["develop"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_component.test", "name", componentName),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_component.test", "slug", componentName),
					resource.TestCheckResourceAttr("instellar_component.test", "current_state", "active"),
					resource.TestCheckResourceAttr("instellar_component.test", "channels.#", "1"),
					// Verify dynamic vlaues have value set
					resource.TestCheckResourceAttrSet("instellar_component.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_component.test", "last_updated"),
				),
			},
			{
				ResourceName:            "instellar_component.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "current_state"},
			},
			{
				Config: buildConfig(clusterNameSlug, componentName, `["develop", "master"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_component.test", "name", componentName),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_component.test", "slug", componentName),
					resource.TestCheckResourceAttr("instellar_component.test", "current_state", "active"),
					resource.TestCheckResourceAttr("instellar_component.test", "channels.#", "2"),
					// Verify dynamic values have value set
					resource.TestCheckResourceAttrSet("instellar_component.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_component.test", "last_updated"),
				),
			},
		},
	})
}

func buildConfig(clusterName string, componentName string, channels string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
		resource "instellar_cluster" "test" {
			name = "%s"
			provider_name = "aws"
			region = "ap-southeast-1"
			endpoint = "127.0.0.1:8443"
			password_token = "some-password-or-token"
		}

		resource "instellar_component" "test" {
			name = "%s"
			provider_name = "aws"
			driver = "database/postgresql"
			driver_version = "15.2"
			cluster_ids = [
				instellar_cluster.test.id
			]
			channels = %s
			credential {
				username = "postgres"
				password = "postgres"
				database = "postgres"
				host = "localhost"
				port = 5432
			}
		}
	`, clusterName, componentName, channels)
}
