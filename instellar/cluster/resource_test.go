package cluster_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/upmaru/terraform-provider-instellar/internal/acceptance"
)

func TestAccClusterResource(t *testing.T) {
	clusterUUID := uuid.New()
	clusterNameSegments := strings.Split(clusterUUID.String(), "-")
	clusterNameSlug := strings.Join([]string{clusterNameSegments[0], clusterNameSegments[1]}, "-")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildConfig(clusterNameSlug, "127.0.0.1:8443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_cluster.test", "name", clusterNameSlug),
					resource.TestCheckResourceAttr("instellar_cluster.test", "endpoint", "127.0.0.1:8443"),

					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_cluster.test", "slug", clusterNameSlug),
					resource.TestCheckResourceAttr("instellar_cluster.test", "current_state", "connecting"),
					// Verify dynamic vlaues have value set
					resource.TestCheckResourceAttrSet("instellar_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_cluster.test", "last_updated"),
				),
			},
			{
				ResourceName:            "instellar_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "password_token", "current_state", "insterra_component_id"},
			},
			{
				Config: buildConfig(clusterNameSlug, "38.43.56.78:8443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_cluster.test", "name", clusterNameSlug),
					resource.TestCheckResourceAttr("instellar_cluster.test", "endpoint", "38.43.56.78:8443"),
					// Verify computed attribute fields.
					resource.TestCheckResourceAttr("instellar_cluster.test", "slug", clusterNameSlug),
					// Verify dynamic values have value set
					resource.TestCheckResourceAttrSet("instellar_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("instellar_cluster.test", "last_updated"),
				),
			},
		},
	})
}

func buildConfig(clusterNameSlug string, endpoint string) string {
	return acceptance.ProviderConfig + fmt.Sprintf(`
		resource "instellar_cluster" "test" {
			name = "%s"
			provider_name = "aws"
			region = "ap-southeast-1"
			endpoint = "%s"
			password_token = "some-password-or-token"
		}
	`, clusterNameSlug, endpoint)
}
