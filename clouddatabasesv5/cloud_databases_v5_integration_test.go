//go:build integration
// +build integration

/**
 * (C) Copyright IBM Corp. 2022.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clouddatabasesv5_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IBM/cloud-databases-go-sdk/clouddatabasesv5"
	"github.com/IBM/go-sdk-core/v5/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

/**
 * This file contains an integration test for the clouddatabasesv5 package.
 *
 * Notes:
 *
 * The integration test will automatically skip tests if the required config file is not available.
 */

var _ = Describe(`CloudDatabasesV5 Integration Tests`, func() {

	const externalConfigFile = "../cloud_databases.env"

	var (
		err                   error
		cloudDatabasesService *clouddatabasesv5.CloudDatabasesV5
		serviceURL            string
		deploymentID          string
		replicaID             string
		autoScalingGroupID    string = "member"
		config                map[string]string
	)

	// Globlal variables to hold link values
	var (
		backupIDLink       string
		scalingGroupIDLink string
		taskIDLink         string
	)

	var shouldSkipTest = func() {
		Skip("External configuration is not available, skipping tests...")
	}

	var waitForTask = func(taskID string) {
		getTaskOptions := &clouddatabasesv5.GetTaskOptions{
			ID: &taskID,
		}

		// If the task runs for more than a minute, then we'll consider it to have succeeded.
		for complete, attempts := false, 0; !complete && attempts < 30; attempts++ {
			getTaskResponse, response, err := cloudDatabasesService.GetTask(getTaskOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(getTaskResponse).ToNot(BeNil())

			if getTaskResponse.Task == nil {
				complete = true
			} else {
				switch *getTaskResponse.Task.Status {
				case "completed", "failed":
					complete = true
					Expect(*getTaskResponse.Task.Status).To(Equal("completed"))
				case "queued", "running":
					break // from switch, not from for
				default:
					fmt.Println("status is " + *getTaskResponse.Task.Status)
				}
			}
			time.Sleep(2 * time.Second)
		}
	}

	Describe(`External configuration`, func() {
		It("Successfully load the configuration", func() {
			_, err = os.Stat(externalConfigFile)
			if err != nil {
				Skip("External configuration file not found, skipping tests: " + err.Error())
			}

			os.Setenv("IBM_CREDENTIALS_FILE", externalConfigFile)
			config, err = core.GetServiceProperties(clouddatabasesv5.DefaultServiceName)
			if err != nil {
				Skip("Error loading service properties, skipping tests: " + err.Error())
			}
			serviceURL = config["URL"]
			if serviceURL == "" {
				Skip("Unable to load service URL configuration property, skipping tests")
			}

			deploymentID = config["DEPLOYMENT_ID"]
			if deploymentID == "" {
				Skip("Unable to load service DEPLOYMENT_ID configuration property, skipping tests")
			}
			replicaID = config["REPLICA_ID"]
			if replicaID == "" {
				Skip("Unable to load service REPLICA_ID configuration property, skipping tests")
			}

			fmt.Printf("Service URL: %s\n", serviceURL)
			shouldSkipTest = func() {}
		})
	})

	Describe(`Client initialization`, func() {
		It("Successfully construct the service client instance", func() {

			cloudDatabasesServiceOptions := &clouddatabasesv5.CloudDatabasesV5Options{}

			cloudDatabasesService, err = clouddatabasesv5.NewCloudDatabasesV5UsingExternalConfig(cloudDatabasesServiceOptions)

			Expect(err).To(BeNil())
			Expect(cloudDatabasesService).ToNot(BeNil())
			Expect(cloudDatabasesService.Service.Options.URL).To(Equal(serviceURL))

			core.SetLogger(core.NewLogger(core.LevelDebug, log.New(GinkgoWriter, "", log.LstdFlags), log.New(GinkgoWriter, "", log.LstdFlags)))
			cloudDatabasesService.EnableRetries(4, 30*time.Second)
		})
	})

	Describe(`AddAllowlistEntry - Add an address or range to the allowlist for a deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`AddAllowlistEntry(addAllowlistEntryOptions *AddAllowlistEntryOptions)`, func() {

			allowlistEntryModel := &clouddatabasesv5.AllowlistEntry{
				Address:     core.StringPtr("172.16.0.0/16"),
				Description: core.StringPtr("Dev IP space 3"),
			}

			addAllowlistEntryOptions := &clouddatabasesv5.AddAllowlistEntryOptions{
				ID:        &deploymentID,
				IPAddress: allowlistEntryModel,
			}

			addAllowlistEntryResponse, response, err := cloudDatabasesService.AddAllowlistEntry(addAllowlistEntryOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(addAllowlistEntryResponse).ToNot(BeNil())

			taskIDLink = *addAllowlistEntryResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`DeleteAllowlistEntry - Delete an address or range from the allowlist of a deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DeleteAllowlistEntry(deleteAllowlistEntryOptions *DeleteAllowlistEntryOptions)`, func() {

			deleteAllowlistEntryOptions := &clouddatabasesv5.DeleteAllowlistEntryOptions{
				ID:        &deploymentID,
				Ipaddress: core.StringPtr("172.16.0.0/16"),
			}

			deleteAllowlistEntryResponse, response, err := cloudDatabasesService.DeleteAllowlistEntry(deleteAllowlistEntryOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(deleteAllowlistEntryResponse).ToNot(BeNil())

			taskIDLink = *deleteAllowlistEntryResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`CreateDatabaseUser - Creates a user based on user type`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CreateDatabaseUser(createDatabaseUserOptions *CreateDatabaseUserOptions)`, func() {

			userModel := &clouddatabasesv5.User{
				Username: core.StringPtr("user"),
				Password: core.StringPtr("password123"),
			}

			createDatabaseUserOptions := &clouddatabasesv5.CreateDatabaseUserOptions{
				ID:       &deploymentID,
				UserType: core.StringPtr("database"),
				User:     userModel,
			}

			createDatabaseUserResponse, response, err := cloudDatabasesService.CreateDatabaseUser(createDatabaseUserOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(createDatabaseUserResponse).ToNot(BeNil())

			taskIDLink = *createDatabaseUserResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`ChangeUserPassword - Set specified user's password`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`ChangeUserPassword(changeUserPasswordOptions *ChangeUserPasswordOptions)`, func() {

			aPasswordSettingUserModel := &clouddatabasesv5.APasswordSettingUser{
				Password: core.StringPtr("xyzzyyzzyx"),
			}

			changeUserPasswordOptions := &clouddatabasesv5.ChangeUserPasswordOptions{
				ID:       &deploymentID,
				UserType: core.StringPtr("database"),
				Username: core.StringPtr("user"),
				User:     aPasswordSettingUserModel,
			}

			changeUserPasswordResponse, response, err := cloudDatabasesService.ChangeUserPassword(changeUserPasswordOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(changeUserPasswordResponse).ToNot(BeNil())

			taskIDLink = *changeUserPasswordResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`DeleteDatabaseUser - Deletes a user based on user type`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DeleteDatabaseUser(deleteDatabaseUserOptions *DeleteDatabaseUserOptions)`, func() {

			deleteDatabaseUserOptions := &clouddatabasesv5.DeleteDatabaseUserOptions{
				ID:       &deploymentID,
				UserType: core.StringPtr("database"),
				Username: core.StringPtr("user"),
			}

			deleteDatabaseUserResponse, response, err := cloudDatabasesService.DeleteDatabaseUser(deleteDatabaseUserOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(deleteDatabaseUserResponse).ToNot(BeNil())

			taskIDLink = *deleteDatabaseUserResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`KillConnections - Kill connections to a PostgreSQL or EnterpriseDB deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`KillConnections(killConnectionsOptions *KillConnectionsOptions)`, func() {

			killConnectionsOptions := &clouddatabasesv5.KillConnectionsOptions{
				ID: &deploymentID,
			}

			killConnectionsResponse, response, err := cloudDatabasesService.KillConnections(killConnectionsOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(killConnectionsResponse).ToNot(BeNil())

			taskIDLink = *killConnectionsResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`SetAllowlist - Set the allowlist for a deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`SetAllowlist(setAllowlistOptions *SetAllowlistOptions)`, func() {

			allowlistEntryModel := &clouddatabasesv5.AllowlistEntry{
				Address:     core.StringPtr("195.212.0.0/16"),
				Description: core.StringPtr("Dev IP space 1"),
			}

			setAllowlistOptions := &clouddatabasesv5.SetAllowlistOptions{
				ID:          &deploymentID,
				IPAddresses: []clouddatabasesv5.AllowlistEntry{*allowlistEntryModel},
			}

			setAllowlistResponse, response, err := cloudDatabasesService.SetAllowlist(setAllowlistOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(setAllowlistResponse).ToNot(BeNil())

			taskIDLink = *setAllowlistResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`SetAutoscalingConditions - Set the autoscaling configuration from a deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`SetAutoscalingConditions(setAutoscalingConditionsOptions *SetAutoscalingConditionsOptions)`, func() {

			autoscalingMemoryGroupMemoryScalersIoUtilizationModel := &clouddatabasesv5.AutoscalingMemoryGroupMemoryScalersIoUtilization{
				Enabled:      core.BoolPtr(true),
				OverPeriod:   core.StringPtr("5m"),
				AbovePercent: core.Int64Ptr(int64(90)),
			}

			autoscalingMemoryGroupMemoryScalersModel := &clouddatabasesv5.AutoscalingMemoryGroupMemoryScalers{
				IoUtilization: autoscalingMemoryGroupMemoryScalersIoUtilizationModel,
			}

			autoscalingMemoryGroupMemoryRateModel := &clouddatabasesv5.AutoscalingMemoryGroupMemoryRate{
				IncreasePercent:  core.Float64Ptr(float64(10.0)),
				PeriodSeconds:    core.Int64Ptr(int64(300)),
				LimitMbPerMember: core.Float64Ptr(float64(114432)),
				Units:            core.StringPtr("mb"),
			}

			autoscalingMemoryGroupMemoryModel := &clouddatabasesv5.AutoscalingMemoryGroupMemory{
				Scalers: autoscalingMemoryGroupMemoryScalersModel,
				Rate:    autoscalingMemoryGroupMemoryRateModel,
			}

			autoscalingSetGroupAutoscalingModel := &clouddatabasesv5.AutoscalingSetGroupAutoscalingAutoscalingMemoryGroup{
				Memory: autoscalingMemoryGroupMemoryModel,
			}

			setAutoscalingConditionsOptions := &clouddatabasesv5.SetAutoscalingConditionsOptions{
				ID:          &deploymentID,
				GroupID:     &autoScalingGroupID,
				Autoscaling: autoscalingSetGroupAutoscalingModel,
			}

			setAutoscalingConditionsResponse, response, err := cloudDatabasesService.SetAutoscalingConditions(setAutoscalingConditionsOptions)

			if err != nil {
				fmt.Printf("\nError: %s", response.Result.(map[string]interface{}))
			}

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(setAutoscalingConditionsResponse).ToNot(BeNil())

			taskIDLink = *setAutoscalingConditionsResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`UpdateDatabaseConfiguration - Change your database configuration`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`UpdateDatabaseConfiguration(updateDatabaseConfigurationOptions *UpdateDatabaseConfigurationOptions)`, func() {

			configurationModel := &clouddatabasesv5.ConfigurationPgConfiguration{
				MaxConnections: core.Int64Ptr(int64(200)),
			}

			updateDatabaseConfigurationOptions := cloudDatabasesService.NewUpdateDatabaseConfigurationOptions(
				deploymentID,
			)
			updateDatabaseConfigurationOptions.SetConfiguration(configurationModel)

			updateDatabaseConfigurationResponse, response, err := cloudDatabasesService.UpdateDatabaseConfiguration(updateDatabaseConfigurationOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(updateDatabaseConfigurationResponse).ToNot(BeNil())

			taskIDLink = *updateDatabaseConfigurationResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`ListDeployables - List all deployable databases`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`ListDeployables(listDeployablesOptions *ListDeployablesOptions)`, func() {

			listDeployablesOptions := &clouddatabasesv5.ListDeployablesOptions{}

			listDeployablesResponse, response, err := cloudDatabasesService.ListDeployables(listDeployablesOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(listDeployablesResponse).ToNot(BeNil())
		})
	})

	Describe(`ListRegions - List all deployable regions`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`ListRegions(listRegionsOptions *ListRegionsOptions)`, func() {

			listRegionsOptions := &clouddatabasesv5.ListRegionsOptions{}

			listRegionsResponse, response, err := cloudDatabasesService.ListRegions(listRegionsOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(listRegionsResponse).ToNot(BeNil())
		})
	})

	Describe(`GetDeploymentInfo - Get deployment information`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetDeploymentInfo(getDeploymentInfoOptions *GetDeploymentInfoOptions)`, func() {

			getDeploymentInfoOptions := &clouddatabasesv5.GetDeploymentInfoOptions{
				ID: &deploymentID,
			}

			getDeploymentInfoResponse, response, err := cloudDatabasesService.GetDeploymentInfo(getDeploymentInfoOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(getDeploymentInfoResponse).ToNot(BeNil())
		})
	})

	Describe(`ListRemotes - List read-only replica information`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`ListRemotes(listRemotesOptions *ListRemotesOptions)`, func() {

			listRemotesOptions := &clouddatabasesv5.ListRemotesOptions{
				ID: &deploymentID,
			}

			listRemotesResponse, response, err := cloudDatabasesService.ListRemotes(listRemotesOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(listRemotesResponse).ToNot(BeNil())
		})
	})

	Describe(`ResyncReplica - Resync read-only replica`, func() {
		BeforeEach(func() {
			if replicaID == "" {
				Skip("Unable to load service REPLICA_ID configuration property, skipping tests")
			}
		})
		It(`ResyncReplica(resyncReplicaOptions *ResyncReplicaOptions)`, func() {

			resyncReplicaOptions := &clouddatabasesv5.ResyncReplicaOptions{
				ID: &replicaID,
			}

			resyncReplicaResponse, response, err := cloudDatabasesService.ResyncReplica(resyncReplicaOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(resyncReplicaResponse).ToNot(BeNil())

			taskIDLink = *resyncReplicaResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`PromoteReadOnlyReplica - Promote read-only replica to a full deployment`, func() {
		BeforeEach(func() {
			if replicaID == "" {
				Skip("Unable to load service REPLICA_ID configuration property, skipping tests")
			}
		})
		It(`PromoteReadOnlyReplica(promoteReadOnlyReplicaOptions *PromoteReadOnlyReplicaOptions)`, func() {

			promoteReadOnlyReplicaOptions := &clouddatabasesv5.PromoteReadOnlyReplicaOptions{
				ID:        &replicaID,
				Promotion: make(map[string]interface{}),
			}

			promoteReadOnlyReplicaResponse, response, err := cloudDatabasesService.PromoteReadOnlyReplica(promoteReadOnlyReplicaOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(promoteReadOnlyReplicaResponse).ToNot(BeNil())

			taskIDLink = *promoteReadOnlyReplicaResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`ListDeploymentTasks - List currently running tasks on a deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`ListDeploymentTasks(listDeploymentTasksOptions *ListDeploymentTasksOptions)`, func() {

			listDeploymentTasksOptions := &clouddatabasesv5.ListDeploymentTasksOptions{
				ID: &deploymentID,
			}

			tasks, response, err := cloudDatabasesService.ListDeploymentTasks(listDeploymentTasksOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(tasks).ToNot(BeNil())
		})
	})

	Describe(`GetTask - Get information about a task`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetTask(getTaskOptions *GetTaskOptions)`, func() {

			getTaskOptions := &clouddatabasesv5.GetTaskOptions{
				ID: &taskIDLink,
			}

			getTaskResponse, response, err := cloudDatabasesService.GetTask(getTaskOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(getTaskResponse).ToNot(BeNil())
		})
	})

	Describe(`ListDeploymentBackups - List currently available backups from a deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`ListDeploymentBackups(listDeploymentBackupsOptions *ListDeploymentBackupsOptions)`, func() {

			listDeploymentBackupsOptions := &clouddatabasesv5.ListDeploymentBackupsOptions{
				ID: &deploymentID,
			}

			backups, response, err := cloudDatabasesService.ListDeploymentBackups(listDeploymentBackupsOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(backups).ToNot(BeNil())

			backupIDLink = *backups.Backups[0].ID
		})
	})

	Describe(`GetBackupInfo - Get information about a backup`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetBackupInfo(getBackupInfoOptions *GetBackupInfoOptions)`, func() {

			getBackupInfoOptions := &clouddatabasesv5.GetBackupInfoOptions{
				BackupID: &backupIDLink,
			}

			getBackupInfoResponse, response, err := cloudDatabasesService.GetBackupInfo(getBackupInfoOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(getBackupInfoResponse).ToNot(BeNil())
		})
	})

	Describe(`StartOndemandBackup - Initiate an on-demand backup`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`StartOndemandBackup(startOndemandBackupOptions *StartOndemandBackupOptions)`, func() {

			startOndemandBackupOptions := &clouddatabasesv5.StartOndemandBackupOptions{
				ID: &deploymentID,
			}

			startOndemandBackupResponse, response, err := cloudDatabasesService.StartOndemandBackup(startOndemandBackupOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(startOndemandBackupResponse).ToNot(BeNil())
		})
	})

	Describe(`GetPitrData - Get earliest point-in-time-recovery timestamp`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetPitrData(getPitrDataOptions *GetPitrDataOptions)`, func() {

			getPitrDataOptions := &clouddatabasesv5.GetPitrDataOptions{
				ID: &deploymentID,
			}

			getPitrDataResponse, response, err := cloudDatabasesService.GetPitrData(getPitrDataOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(getPitrDataResponse.PointInTimeRecoveryData).ToNot(BeNil())
		})
	})

	Describe(`GetConnection - Discover connection information for a deployment for a user with an endpoint type`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetConnection(getConnectionOptions *GetConnectionOptions)`, func() {

			getConnectionOptions := &clouddatabasesv5.GetConnectionOptions{
				ID:              &deploymentID,
				UserType:        core.StringPtr("database"),
				UserID:          core.StringPtr("testuser"),
				EndpointType:    core.StringPtr("public"),
				CertificateRoot: core.StringPtr("/var/test"),
			}

			getConnectionResponse, response, err := cloudDatabasesService.GetConnection(getConnectionOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(getConnectionResponse).ToNot(BeNil())
		})
	})

	Describe(`CompleteConnection - Discover connection information for a deployment for a user with substitutions and an endpoint type`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CompleteConnection(completeConnectionOptions *CompleteConnectionOptions)`, func() {

			completeConnectionOptions := &clouddatabasesv5.CompleteConnectionOptions{
				ID:              &deploymentID,
				UserType:        core.StringPtr("database"),
				UserID:          core.StringPtr("testuser"),
				EndpointType:    core.StringPtr("public"),
				Password:        core.StringPtr("providedpassword"),
				CertificateRoot: core.StringPtr("/var/test"),
			}

			completeConnectionResponse, response, err := cloudDatabasesService.CompleteConnection(completeConnectionOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(completeConnectionResponse).ToNot(BeNil())
		})
	})

	Describe(`ListDeploymentScalingGroups - List currently available scaling groups from a deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`ListDeploymentScalingGroups(listDeploymentScalingGroupsOptions *ListDeploymentScalingGroupsOptions)`, func() {

			listDeploymentScalingGroupsOptions := &clouddatabasesv5.ListDeploymentScalingGroupsOptions{
				ID: &deploymentID,
			}

			listDeploymentScalingGroupsResponse, response, err := cloudDatabasesService.ListDeploymentScalingGroups(listDeploymentScalingGroupsOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(listDeploymentScalingGroupsResponse).ToNot(BeNil())

			scalingGroupIDLink = *listDeploymentScalingGroupsResponse.Groups[0].ID
		})
	})

	Describe(`SetDeploymentScalingGroup - Set scaling values on a specified group`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`SetDeploymentScalingGroup(setDeploymentScalingGroupOptions *SetDeploymentScalingGroupOptions)`, func() {

			groupScalingMemoryModel := &clouddatabasesv5.GroupScalingMemory{
				AllocationMb: core.Int64Ptr(int64(12288)),
			}

			groupScalingDiskModel := &clouddatabasesv5.GroupScalingDisk{
				AllocationMb: core.Int64Ptr(int64(20480)),
			}

			groupScalingModel := &clouddatabasesv5.GroupScaling{
				Memory: groupScalingMemoryModel,
				Disk:   groupScalingDiskModel,
			}

			setDeploymentScalingGroupOptions := &clouddatabasesv5.SetDeploymentScalingGroupOptions{
				ID:      &deploymentID,
				GroupID: &scalingGroupIDLink,
				Group:   groupScalingModel,
			}

			setDeploymentScalingGroupResponse, response, err := cloudDatabasesService.SetDeploymentScalingGroup(setDeploymentScalingGroupOptions)

			if err != nil {
				fmt.Printf("\nError: %s", response.Result.(map[string]interface{}))
			}

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(setDeploymentScalingGroupResponse).ToNot(BeNil())

			taskIDLink = *setDeploymentScalingGroupResponse.Task.ID

			waitForTask(taskIDLink)
		})
	})

	Describe(`GetDefaultScalingGroups - Get default scaling groups for a new deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetDefaultScalingGroups(getDefaultScalingGroupsOptions *GetDefaultScalingGroupsOptions)`, func() {

			getDefaultScalingGroupsOptions := &clouddatabasesv5.GetDefaultScalingGroupsOptions{
				Type: core.StringPtr("postgresql"),
			}

			getDefaultScalingGroupsResponse, response, err := cloudDatabasesService.GetDefaultScalingGroups(getDefaultScalingGroupsOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(getDefaultScalingGroupsResponse).ToNot(BeNil())
		})
	})

	Describe(`GetAutoscalingConditions - Get the autoscaling configuration from a deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetAutoscalingConditions(getAutoscalingConditionsOptions *GetAutoscalingConditionsOptions)`, func() {

			getAutoscalingConditionsOptions := &clouddatabasesv5.GetAutoscalingConditionsOptions{
				ID:      &deploymentID,
				GroupID: &autoScalingGroupID,
			}

			autoscalingGroup, response, err := cloudDatabasesService.GetAutoscalingConditions(getAutoscalingConditionsOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(autoscalingGroup).ToNot(BeNil())
		})
	})

	Describe(`GetAllowlist - Retrieve the allowlisted addresses and ranges for a deployment`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetAllowlist(getAllowlistOptions *GetAllowlistOptions)`, func() {

			getAllowlistOptions := &clouddatabasesv5.GetAllowlistOptions{
				ID: &deploymentID,
			}

			getAllowlistResponse, response, err := cloudDatabasesService.GetAllowlist(getAllowlistOptions)

			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(getAllowlistResponse).ToNot(BeNil())
		})
	})
})
