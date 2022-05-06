package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTFEOrganizationRunTask() *schema.Resource {
	return &schema.Resource{

		Create: resourceTFEOrganizationRunTaskCreate,
		Read:   resourceTFEOrganizationRunTaskRead,
		Delete: resourceTFEOrganizationRunTaskDelete,
		Update: resourceTFEOrganizationRunTaskUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceTFEOrganizationRunTaskImporter,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"url": {
				Type:     schema.TypeString,
				Required: true,
			},

			"category": {
				Type:     schema.TypeString,
				Default:  "task",
				Optional: true,
			},

			"hmac_key": {
				Type:      schema.TypeString,
				Sensitive: true,
				Default:   "",
				Optional:  true,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
		},
	}
}

func resourceTFEOrganizationRunTaskCreate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the task name and organization.
	name := d.Get("name").(string)
	organization := d.Get("organization").(string)

	// Create a new options struct.
	options := tfe.RunTaskCreateOptions{
		Name:     name,
		URL:      d.Get("url").(string),
		Category: d.Get("category").(string),
		HMACKey:  tfe.String(d.Get("hmac_key").(string)),
		Enabled:  tfe.Bool(d.Get("enabled").(bool)),
	}

	log.Printf("[DEBUG] Create task %s for organization: %s", name, organization)
	task, err := tfeClient.RunTasks.Create(ctx, organization, options)
	if err != nil {
		return fmt.Errorf(
			"Error creating task %s for organization %s: %v", name, organization, err)
	}

	d.SetId(task.ID)

	return resourceTFEOrganizationRunTaskRead(d, meta)
}

func resourceTFEOrganizationRunTaskDelete(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete task: %s", d.Id())
	err := tfeClient.RunTasks.Delete(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting task %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFEOrganizationRunTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Setup the options struct
	options := tfe.RunTaskUpdateOptions{}
	if d.HasChange("name") {
		options.Name = tfe.String(d.Get("name").(string))
	}
	if d.HasChange("url") {
		options.URL = tfe.String(d.Get("url").(string))
	}
	if d.HasChange("category") {
		options.Category = tfe.String(d.Get("category").(string))
	}
	if d.HasChange("enabled") {
		options.Enabled = tfe.Bool(d.Get("enabled").(bool))
	}
	if d.HasChange("hmac_key") {
		options.HMACKey = tfe.String(d.Get("hmac_key").(string))
	}

	log.Printf("[DEBUG] Update configuration of task: %s", d.Id())
	task, err := tfeClient.RunTasks.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf("Error updating task %s: %v", d.Id(), err)
	}

	d.SetId(task.ID)

	return resourceTFEOrganizationRunTaskRead(d, meta)
}

func resourceTFEOrganizationRunTaskRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of task: %s", d.Id())
	task, err := tfeClient.RunTasks.Read(ctx, d.Id())

	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Task %s does not exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of task %s: %v", d.Id(), err)
	}

	// Update the config.
	d.Set("name", task.Name)
	d.Set("url", task.URL)
	d.Set("category", task.Category)
	d.Set("enabled", task.Enabled)
	// The HMAC Key is always empty from the API so all we can do is
	// echo the request's key to the response
	d.Set("hmac_key", tfe.String(d.Get("hmac_key").(string)))

	return nil
}

func resourceTFEOrganizationRunTaskImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tfeClient := meta.(*tfe.Client)

	s := strings.Split(d.Id(), "/")
	if len(s) != 2 {
		return nil, fmt.Errorf(
			"invalid task input format: %s (expected <ORGANIZATION>/<TASK NAME>)",
			d.Id(),
		)
	}

	task, err := fetchOrganizationRunTask(s[1], s[0], tfeClient)
	if err != nil {
		return nil, err
	}

	d.Set("organization", task.Organization.Name)
	d.SetId(task.ID)

	return []*schema.ResourceData{d}, nil
}
