package tests

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/gomega"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

// CreateOrganization creates org on px-backup
func CreateOrganization(orgID string) {
	Step(fmt.Sprintf("Create organization [%s]", orgID), func() {
		backupDriver := Inst().Backup
		req := &api.OrganizationCreateRequest{
			CreateMetadata: &api.CreateMetadata{
				Name: orgID,
			},
		}
		_, err := backupDriver.CreateOrganization(req)
		Expect(err).NotTo(HaveOccurred(),
			fmt.Sprintf("Failed to create organization [%s]", orgID))
	})
}

// TODO: There is no delete org API
/*func DeleteOrganization(orgID string) {
	Step(fmt.Sprintf("Delete organization [%s]", orgID), func() {
		backupDriver := Inst().Backup
		req := &api.Delete{
			CreateMetadata: &api.CreateMetadata{
				Name: orgID,
			},
		}
		_, err := backupDriver.Delete(req)
		Expect(err).NotTo(HaveOccurred())
	})
}*/

// DeleteCloudCredential deletes cloud credentials
func DeleteCloudCredential(name string, orgID string) {
	Step(fmt.Sprintf("Delete cloud credential [%s] in org [%s]", name, orgID), func() {
		backupDriver := Inst().Backup

		credDeleteRequest := &api.CloudCredentialDeleteRequest{
			Name:  name,
			OrgId: orgID,
		}
		backupDriver.DeleteCloudCredential(credDeleteRequest)
		// Best effort cleanup, dont fail test, if deletion fails
		// Expect(err).NotTo(HaveOccurred(),
		//	fmt.Sprintf("Failed to delete cloud credential [%s] in org [%s]", name, orgID))
		// TODO: validate CreateCloudCredentialResponse also
	})

}

// CreateCloudCredential creates cloud credetials
func CreateCloudCredential(name string, orgID string) {

	Step(fmt.Sprintf("Create cloud credential [%s] in org [%s]", name, orgID), func() {
		backupDriver := Inst().Backup

		// TODO: add separate function to return cred object based on type
		id := os.Getenv("AWS_ACCESS_KEY_ID")
		Expect(id).NotTo(Equal(""),
			"AWS_ACCESS_KEY_ID Environment variable should not be empty")

		secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
		Expect(secret).NotTo(Equal(""),
			"AWS_SECRET_ACCESS_KEY Environment variable should not be empty")

		credCreateRequest := &api.CloudCredentialCreateRequest{
			CreateMetadata: &api.CreateMetadata{
				Name:  name,
				OrgId: orgID,
			},
			CloudCredential: &api.CloudCredentialInfo{
				Type: api.CloudCredentialInfo_AWS,
				Config: &api.CloudCredentialInfo_AwsConfig{
					AwsConfig: &api.AWSConfig{
						AccessKey: id,
						SecretKey: secret,
					},
				},
			},
		}
		_, err := backupDriver.CreateCloudCredential(credCreateRequest)
		Expect(err).NotTo(HaveOccurred(),
			fmt.Sprintf("Failed to create cloud credential [%s] in org [%s]", name, orgID))
		// TODO: validate CreateCloudCredentialResponse also
	})

}

// CreateBackupLocation creates backup location
func CreateBackupLocation(name string, cloudCred string,
	bLocationType api.BackupLocationInfo_Type, orgID string) {

	Step(fmt.Sprintf("Create backup location [%s] in org [%s]", name, orgID), func() {

		switch bLocationType {
		case api.BackupLocationInfo_S3:
			CreateS3BackupLocation(name, cloudCred, orgID)
		case api.BackupLocationInfo_Azure, api.BackupLocationInfo_Google:
			// TODO add support for other platforms
			return
		default:
			return
		}
	})
}

// CreateS3BackupLocation creates backuplocation for S3
func CreateS3BackupLocation(name string, cloudCred string, orgID string) {
	backupDriver := Inst().Backup
	path := os.Getenv("BUCKET_NAME")
	Expect(path).NotTo(Equal(""),
		"BUCKET_NAME Environment variable should not be empty")

	endpoint := os.Getenv("S3_ENDPOINT")
	Expect(endpoint).NotTo(Equal(""),
		"S3_ENDPOINT Environment variable should not be empty")

	region := os.Getenv("S3_REGION")
	Expect(endpoint).NotTo(Equal(""),
		"S3_REGION Environment variable should not be empty")

	disableSSL := os.Getenv("S3_DISABLE_SSL")
	Expect(endpoint).NotTo(Equal(""),
		"S3_DISABLE_SSL Environment variable should not be empty")
	disableSSLBool, err := strconv.ParseBool(disableSSL)
	Expect(err).NotTo(HaveOccurred(),
		fmt.Sprintf("S3_DISABLE_SSL=%s is not a valid boolean value", disableSSL))

	encryptionKey := "torpedo"
	bLocationCreateReq := &api.BackupLocationCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  name,
			OrgId: orgID,
		},
		BackupLocation: &api.BackupLocationInfo{
			Path:            path,
			EncryptionKey:   encryptionKey,
			CloudCredential: cloudCred,
			Type:            api.BackupLocationInfo_S3,
			Config: &api.BackupLocationInfo_S3Config{
				S3Config: &api.S3Config{
					Endpoint:   endpoint,
					Region:     region,
					DisableSsl: disableSSLBool,
				},
			},
		},
	}
	_, err = backupDriver.CreateBackupLocation(bLocationCreateReq)
	Expect(err).NotTo(HaveOccurred(),
		fmt.Sprintf("Failed to create backuplocation [%s] in org [%s]", name, orgID))
}

// DeleteBackupLocation deletes backuplocation
func DeleteBackupLocation(name string, orgID string) {
	Step(fmt.Sprintf("Delete backup location [%s] in org [%s]", name, orgID), func() {
		backupDriver := Inst().Backup
		bLocationDeleteReq := &api.BackupLocationDeleteRequest{
			Name:  name,
			OrgId: orgID,
		}
		backupDriver.DeleteBackupLocation(bLocationDeleteReq)
		// Best effort cleanup, dont fail test, if deletion fails
		//Expect(err).NotTo(HaveOccurred(),
		//	fmt.Sprintf("Failed to delete backup location [%s] in org [%s]", name, orgID))
		// TODO: validate createBackupLocationResponse also
	})
}

// CreateSourceAndDestClusters creates source and destination cluster
// 1st cluster in KUBECONFIGS ENV var is source cluster while
// 2nd cluster is destination cluster
func CreateSourceAndDestClusters(cloudCred string, orgID string) {
	// TODO: Add support for adding multiple clusters from
	// comma separated list of kubeconfig files
	kubeconfigs := os.Getenv("KUBECONFIGS")
	Expect(kubeconfigs).NotTo(Equal(""),
		"KUBECONFIGS Environment variable should not be empty")

	kubeconfigList := strings.Split(kubeconfigs, ",")
	// Validate user has provided at least 2 kubeconfigs for source and destination cluster
	Expect(len(kubeconfigList)).Should(BeNumerically(">=", 2), "At least minimum two kubeconfigs required")

	err := dumpKubeConfigs(ConfigMapName, kubeconfigList)
	Expect(err).NotTo(HaveOccurred(),
		fmt.Sprintf("Failed to get kubeconfigs [%v] from configmap [%s]", kubeconfigList, ConfigMapName))

	// Register source cluster with backup driver
	Step(fmt.Sprintf("Create cluster [%s] in org [%s]", SourceClusterName, orgID), func() {
		srcClusterKubeConfig := fmt.Sprintf("%s/%s", KubeconfigDirectory, kubeconfigList[0])
		CreateCluster(SourceClusterName, cloudCred, srcClusterKubeConfig, orgID)
	})

	// Register destination cluster with backup driver
	Step(fmt.Sprintf("Create cluster [%s] in org [%s]", DestinationClusterName, orgID), func() {
		destClusterKubeConfig := fmt.Sprintf("%s/%s", KubeconfigDirectory, kubeconfigList[1])
		CreateCluster(DestinationClusterName, cloudCred, destClusterKubeConfig, orgID)
	})
}

func dumpKubeConfigs(configObject string, kubeconfigList []string) error {
	cm, err := core.Instance().GetConfigMap(configObject, "default")
	if err != nil {
		logrus.Errorf("Error reading config map: %v", err)
		return err
	}
	for _, kubeconfig := range kubeconfigList {

		config := cm.Data[kubeconfig]
		if len(config) == 0 {
			configErr := fmt.Sprintf("Error reading kubeconfig: found empty %s in config map %s",
				kubeconfig, configObject)
			return fmt.Errorf(configErr)
		}
		filePath := fmt.Sprintf("%s/%s", KubeconfigDirectory, kubeconfig)
		err := ioutil.WriteFile(filePath, []byte(config), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteCluster deletes/de-registers cluster from px-backup
func DeleteCluster(name string, orgID string) {

	Step(fmt.Sprintf("Delete cluster [%s] in org [%s]", name, orgID), func() {
		backupDriver := Inst().Backup
		clusterDeleteReq := &api.ClusterDeleteRequest{
			OrgId: orgID,
			Name:  name,
		}
		backupDriver.DeleteCluster(clusterDeleteReq)
		// Best effort cleanup, dont fail test, if deletion fails
		//Expect(err).NotTo(HaveOccurred(),
		//	fmt.Sprintf("Failed to delete cluster [%s] in org [%s]", name, orgID))
	})
}

// CreateCluster creates/registers cluster with px-backup
func CreateCluster(name string, cloudCred string, kubeconfigPath string, orgID string) {

	Step(fmt.Sprintf("Create cluster [%s] in org [%s]", name, orgID), func() {
		backupDriver := Inst().Backup
		kubeconfigRaw, err := ioutil.ReadFile(kubeconfigPath)
		Expect(err).NotTo(HaveOccurred(),
			fmt.Sprintf("Failed to read kubeconfig file from location [%s]. Error:[%v]",
				kubeconfigPath, err))
		clusterCreateReq := &api.ClusterCreateRequest{
			CreateMetadata: &api.CreateMetadata{
				Name:  name,
				OrgId: orgID,
			},
			Cluster: &api.ClusterInfo{
				Kubeconfig:      base64.StdEncoding.EncodeToString(kubeconfigRaw),
				CloudCredential: cloudCred,
			},
		}
		_, err = backupDriver.CreateCluster(clusterCreateReq)
		Expect(err).NotTo(HaveOccurred(),
			fmt.Sprintf("Failed to create cluster [%s] in org [%s]. Error : [%v]",
				name, orgID, err))
	})
}

// CreateBackup creates backup
func CreateBackup(backupName string, clusterName string, bLocation string,
	namespaces []string, labelSelectores map[string]string, orgID string) {

	Step(fmt.Sprintf("Create backup [%s] in org [%s] from cluster [%s]",
		backupName, orgID, clusterName), func() {

		backupDriver := Inst().Backup
		bkpCreateRequest := &api.BackupCreateRequest{
			CreateMetadata: &api.CreateMetadata{
				Name:  backupName,
				OrgId: orgID,
			},
			BackupLocation: bLocation,
			Cluster:        clusterName,
			Namespaces:     namespaces,
			LabelSelectors: labelSelectores,
		}
		_, err := backupDriver.CreateBackup(bkpCreateRequest)
		Expect(err).NotTo(HaveOccurred(),
			fmt.Sprintf("Failed to create backup [%s] in org [%s]", backupName, orgID))
		// TODO: validate createClusterResponse also

	})
}

// CreateRestore creates restore
func CreateRestore(restoreName string, backupName string,
	namespaceMapping map[string]string, clusterName string, orgID string) {

	Step(fmt.Sprintf("Create restore [%s] in org [%s] on cluster [%s]",
		restoreName, orgID, clusterName), func() {

		backupDriver := Inst().Backup
		createRestoreReq := &api.RestoreCreateRequest{
			CreateMetadata: &api.CreateMetadata{
				Name:  restoreName,
				OrgId: orgID,
			},
			Backup:           backupName,
			Cluster:          clusterName,
			NamespaceMapping: namespaceMapping,
		}
		_, err := backupDriver.CreateRestore(createRestoreReq)
		Expect(err).NotTo(HaveOccurred(),
			fmt.Sprintf("Failed to create restore [%s] in org [%s] on cluster [%s]",
				restoreName, orgID, clusterName))
		// TODO: validate createClusterResponse also
	})
}

// DeleteBackup deletes backup
func DeleteBackup(backupName string, clusterName string, orgID string) {

	Step(fmt.Sprintf("Delete backup [%s] in org [%s] from cluster [%s]",
		backupName, orgID, clusterName), func() {

		backupDriver := Inst().Backup
		bkpDeleteRequest := &api.BackupDeleteRequest{
			Name:  backupName,
			OrgId: orgID,
		}
		backupDriver.DeleteBackup(bkpDeleteRequest)
		// Best effort cleanup, dont fail test, if deletion fails
		//Expect(err).NotTo(HaveOccurred(),
		//	fmt.Sprintf("Failed to delete backup [%s] in org [%s]", backupName, orgID))
		// TODO: validate createClusterResponse also
	})
}