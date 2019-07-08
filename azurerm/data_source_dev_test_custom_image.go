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

			// There's a bug in the Azure API where this is returned in lower-case
			// BUG: https://github.com/Azure/azure-rest-api-specs/issues/3964
			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"author": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
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

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	d.Set("lab_name", labName)
	d.Set("location", resp.Location)
	d.Set("author", resp.Author)
	d.Set("description", resp.Description)

	return nil
}
