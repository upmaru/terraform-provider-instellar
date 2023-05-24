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
	clusterNameSlug := clusterNameSegments[0]

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acceptance.ProviderConfig + fmt.Sprintf(`
					resource "instellar_cluster" "test" {
						name = "%s"
						provider_name = "aws"
						region = "ap-southeast-1"
						endpoint = "127.0.0.1:8443"
						password_token = "some-password-or-token"
					}
				`, clusterNameSlug),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("instellar_cluster.test", "name", clusterNameSlug),
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
				ImportStateVerifyIgnore: []string{"last_updated", "password_token", "current_state"},
			},
		},
	})
}
