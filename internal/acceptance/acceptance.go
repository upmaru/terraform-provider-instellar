package acceptance

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/upmaru/terraform-provider-instellar/instellar"
)

const (
	ProviderConfig = `
		provider "instellar" {}
	`
)

var (
	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"instellar": providerserver.NewProtocol6WithError(instellar.New()),
	}
)
