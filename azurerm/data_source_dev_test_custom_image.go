package azurerm

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceArmDevTestCustomImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmDevTestCustomImageRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"lab_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.DevTestLabName(),
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),
		},
	}
}

func dataSourceArmDevTestCustomImageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).devTestLabs.CustomImagesClient
	ctx := meta.(*ArmClient).StopContext

	resGroup := d.Get("resource_group_name").(string)
	labName := d.Get("lab_name").(string)
	name := d.Get("name").(string)

	resp, err := client.Get(ctx, resGroup, labName, name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Error: Custom Image %q in Dev Test Lab %q (Resource Group %q) was not found", name, labName, resGroup)
		}

		return fmt.Errorf("Error making Read request on Custom Image %q in Dev Test Lab %q (Resource Group %q): %+v", name, labName, resGroup, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("API returns a nil/empty id on Custom Image %q in Dev Test Lab %q (Resource Group %q): %+v", name, labName, resGroup, err)
	}
	d.SetId(*resp.ID)

	return nil
}
