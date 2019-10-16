package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/devtestlabs/mgmt/2016-05-15/dtl"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmDevTestCustomImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmDevTestCustomImageCreateUpdate,
		Read:   resourceArmDevTestCustomImageRead,
		Update: resourceArmDevTestCustomImageCreateUpdate,
		Delete: resourceArmDevTestCustomImageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"lab_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.DevTestLabName(),
			},

			"location": azure.SchemaLocation(),

			// There's a bug in the Azure API where this is returned in lower-case
			// BUG: https://github.com/Azure/azure-rest-api-specs/issues/3964
			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"author": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"vm": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_vm_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"linux_os_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"linux_os_state": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validate.DevTestCustomImageLinuxOsStateType(),
									},
								},
							},
						},

						"windows_os_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"windows_os_state": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validate.DevTestCustomImageWindowsOsStateType(),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceArmDevTestCustomImageCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).devTestLabs.CustomImagesClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for DevTest Custom Image creation")

	name := d.Get("name").(string)
	labName := d.Get("lab_name").(string)
	resGroup := d.Get("resource_group_name").(string)

	if requireResourcesToBeImported && d.IsNewResource() {
		existing, err := client.Get(ctx, resGroup, labName, name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Custom Image %q in Dev Test %q (Resource Group %q): %s", name, labName, resGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_dev_test_custom_image", *existing.ID)
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	author := d.Get("author").(string)
	description := d.Get("description").(string)

	vmRaw := d.Get("vm").([]interface{})
	vm := expandDevTestCustomImageVMProperties(vmRaw)

	properties := dtl.CustomImageProperties{
		Author:      utils.String(author),
		Description: utils.String(description),
		VM:          vm,
	}

	parameters := dtl.CustomImage{
		Location:              utils.String(location),
		CustomImageProperties: &properties,
	}

	future, err := client.CreateOrUpdate(ctx, resGroup, labName, name, parameters)
	if err != nil {
		return fmt.Errorf("Error creating/updating Custom Image %q in DevTest Lab %q (Resource Group %q): %+v", name, labName, resGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for creation/update of Custom Image %q in DevTest Lab %q (Resource Group %q): %+v", name, labName, resGroup, err)
	}

	read, err := client.Get(ctx, resGroup, labName, name, "")
	if err != nil {
		return fmt.Errorf("Error retrieving Custom Image %q in DevTest Lab %q (Resource Group %q): %+v", name, labName, resGroup, err)
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Custom Image %q in DevTest Lab %q (Resource Group %q) ID", name, labName, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmDevTestCustomImageRead(d, meta)
}

func resourceArmDevTestCustomImageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).devTestLabs.CustomImagesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	labName := id.Path["labs"]
	name := id.Path["customimages"]

	read, err := client.Get(ctx, resGroup, labName, name, "")
	if err != nil {
		if utils.ResponseWasNotFound(read.Response) {
			log.Printf("[DEBUG] Custom Image %q in DevTest Lab %q was not found in Resource Group %q - removing from state!", name, labName, resGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Custom Image %q in DevTest Lab %q (Resource Group %q): %+v", name, labName, resGroup, err)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resGroup)
	if location := read.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := read.CustomImageProperties; props != nil {

		if author := props.Author; author != nil {
			d.Set("author", string(*author))
		}

		if description := props.Description; description != nil {
			d.Set("description", string(*description))
		}

		if vmProps := props.VM; vmProps != nil {
			flattenedVMProps := flattenDevTestCustomImageVMProperties(props.VM)
			if err := d.Set("vm", flattenedVMProps); err != nil {
				return fmt.Errorf("Error setting `vm`: %+v", err)
			}
		}
	}

	return nil
}

func resourceArmDevTestCustomImageDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).devTestLabs.CustomImagesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resGroup := id.ResourceGroup
	labName := id.Path["labs"]
	name := id.Path["customimages"]

	read, err := client.Get(ctx, resGroup, labName, name, "")
	if err != nil {
		if utils.ResponseWasNotFound(read.Response) {
			// deleted outside of TF
			log.Printf("[DEBUG] Custom Image %q in DevTest Lab %q was not found in Resource Group %q - assuming removed!", name, labName, resGroup)
			return nil
		}

		return fmt.Errorf("Error retrieving Custom Image %q in DevTest Lab %q (Resource Group %q): %+v", name, labName, resGroup, err)
	}

	future, err := client.Delete(ctx, resGroup, labName, name)
	if err != nil {
		return fmt.Errorf("Error deleting Custom Image %q in DevTest Lab %q (Resource Group %q): %+v", name, labName, resGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for the deletion of Custom Image %q in DevTest Lab %q (Resource Group %q): %+v", name, labName, resGroup, err)
	}

	return err
}

func expandDevTestCustomImageVMProperties(input []interface{}) *dtl.CustomImagePropertiesFromVM {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	sourceVMID := v["source_vm_id"].(string)

	linuxOsInfoRaw := v["linux_os_info"].([]interface{})
	linuxOsInfo := expandDevTestCustomImageLinuxOsInfo(linuxOsInfoRaw)

	windowsOsInfoRaw := v["windows_os_info"].([]interface{})
	windowsOsInfo := expandDevTestCustomImageWindowsOsInfo(windowsOsInfoRaw)

	vm := dtl.CustomImagePropertiesFromVM{
		SourceVMID:    utils.String(sourceVMID),
		LinuxOsInfo:   linuxOsInfo,
		WindowsOsInfo: windowsOsInfo,
	}

	return &vm
}

func expandDevTestCustomImageLinuxOsInfo(input []interface{}) *dtl.LinuxOsInfo {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})
	linuxOsState := v["linux_os_state"]

	if linuxOsState.(string) == "" {
		return nil
	}

	return &dtl.LinuxOsInfo{
		LinuxOsState: dtl.LinuxOsState(linuxOsState.(string)),
	}
}

func expandDevTestCustomImageWindowsOsInfo(input []interface{}) *dtl.WindowsOsInfo {
	if len(input) == 0 {
		return nil
	}

	v := input[0].(map[string]interface{})
	windowsOsState := v["windows_os_state"]

	if windowsOsState.(string) == "" {
		return nil
	}

	return &dtl.WindowsOsInfo{
		WindowsOsState: dtl.WindowsOsState(windowsOsState.(string)),
	}
}

func flattenDevTestCustomImageVMProperties(input *dtl.CustomImagePropertiesFromVM) []interface{} {
	outputs := make([]interface{}, 0)
	if input == nil {
		return outputs
	}

	properties := make(map[string]interface{})
	if input.SourceVMID != nil {
		properties["source_vm_id"] = *input.SourceVMID
	}

	if input.LinuxOsInfo != nil {
		properties["linux_os_info"] = flattenDevTestCustomImageLinuxOsInfo(input.LinuxOsInfo)
	}

	if input.WindowsOsInfo != nil {
		properties["windows_os_info"] = flattenDevTestCustomImageWindowsOsInfo(input.WindowsOsInfo)
	}

	outputs = append(outputs, properties)

	return outputs
}

func flattenDevTestCustomImageLinuxOsInfo(input *dtl.LinuxOsInfo) []interface{} {
	outputs := make([]interface{}, 0)
	if input == nil {
		return outputs
	}

	info := make(map[string]interface{})
	info["linux_os_state"] = string(input.LinuxOsState)

	outputs = append(outputs, info)

	return outputs
}

func flattenDevTestCustomImageWindowsOsInfo(input *dtl.WindowsOsInfo) []interface{} {
	outputs := make([]interface{}, 0)
	if input == nil {
		return outputs
	}

	info := make(map[string]interface{})
	info["windows_os_state"] = string(input.WindowsOsState)

	outputs = append(outputs, info)

	return outputs
}
