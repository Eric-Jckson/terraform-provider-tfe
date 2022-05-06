package tfe

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTFEOrganizationRunTask_create(t *testing.T) {
	skipUnlessRunTasksDefined(t)
	skipIfFreeOnly(t) // Run Tasks requires TFE or a TFC paid/trial subscription

	runTask := &tfe.RunTask{}
	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFEOrganizationRunTaskDestroy,
		Steps: []resource.TestStep{
			testCheckCreateOrgWithRunTasks(orgName),
			{
				Config: testAccTFEOrganizationRunTask_basic(orgName, rInt, runTasksUrl()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTFEOrganizationRunTaskExists("tfe_organization_run_task.foobar", runTask),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "name", fmt.Sprintf("foobar-task-%d", rInt)),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "url", runTasksUrl()),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "category", "task"),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "hmac_key", ""),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "enabled", "false"),
				),
			},
			{
				Config: testAccTFEOrganizationRunTask_update(orgName, rInt, runTasksUrl()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "name", fmt.Sprintf("foobar-task-%d-new", rInt)),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "url", runTasksUrl()),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "category", "task"),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "hmac_key", "somepassword"),
					resource.TestCheckResourceAttr("tfe_organization_run_task.foobar", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccTFEOrganizationRunTask_import(t *testing.T) {
	skipUnlessRunTasksDefined(t)
	skipIfFreeOnly(t) // Run Tasks requires TFE or a TFC paid/trial subscription

	rInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int()
	orgName := fmt.Sprintf("tst-terraform-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTFETeamAccessDestroy,
		Steps: []resource.TestStep{
			testCheckCreateOrgWithRunTasks(orgName),
			{
				Config: testAccTFEOrganizationRunTask_basic(orgName, rInt, runTasksUrl()),
			},
			{
				ResourceName:      "tfe_organization_run_task.foobar",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("tst-terraform-%d/foobar-task-%d", rInt, rInt),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTFEOrganizationRunTaskExists(n string, runTask *tfe.RunTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tfeClient := testAccProvider.Meta().(*tfe.Client)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}
		rt, err := tfeClient.RunTasks.Read(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if rt == nil {
			return fmt.Errorf("Organization Run Task not found")
		}

		*runTask = *rt

		return nil
	}
}

func testAccCheckTFEOrganizationRunTaskDestroy(s *terraform.State) error {
	tfeClient := testAccProvider.Meta().(*tfe.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tfe_organization_run_task" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, err := tfeClient.RunTasks.Read(ctx, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Organization Run Task %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccTFEOrganizationRunTask_basic(orgName string, rInt int, runTaskUrl string) string {
	return fmt.Sprintf(`
resource "tfe_organization" "foobar" {
	name  = "%s"
	email = "admin@company.com"
}

resource "tfe_organization_run_task" "foobar" {
	organization = tfe_organization.foobar.id
	url          = "%s"
	name         = "foobar-task-%d"
	enabled      = false
}
`, orgName, runTaskUrl, rInt)
}

func testAccTFEOrganizationRunTask_update(orgName string, rInt int, runTaskUrl string) string {
	return fmt.Sprintf(`
	resource "tfe_organization" "foobar" {
		name  = "%s"
		email = "admin@company.com"
	}

	resource "tfe_organization_run_task" "foobar" {
		organization = tfe_organization.foobar.id
		url          = "%s"
		name         = "foobar-task-%d-new"
		enabled      = true
		hmac_key     = "somepassword"
	}
`, orgName, runTaskUrl, rInt)
}
