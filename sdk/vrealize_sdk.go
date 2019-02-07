package sdk

import (
	"fmt"
	"strconv"
	"strings"

	logging "github.com/op/go-logging"
	"github.com/vmware/terraform-provider-vra7/client"
	"github.com/vmware/terraform-provider-vra7/utils"
)

var (
	log = logging.MustGetLogger(utils.LoggerID)
)

//GetCatalogItemRequestTemplate - Call to retrieve a request template for a catalog item.
func GetCatalogItemRequestTemplate(catalogItemID string) (*utils.CatalogItemRequestTemplate, error) {

	//Form a path to read catalog request template via REST call
	path := fmt.Sprintf("/catalog-service/api/consumer/entitledCatalogItems/"+
		"%s/requests/template",
		catalogItemID)
	url := client.BuildEncodedURL(path, nil)
	resp, respErr := client.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var requestTemplate utils.CatalogItemRequestTemplate
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &requestTemplate)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &requestTemplate, nil
}

// ReadCatalogItemNameByID - This function returns the catalog item name using catalog item ID
func ReadCatalogItemNameByID(catalogItemID string) (string, error) {

	path := fmt.Sprintf("/catalog-service/api/consumer/entitledCatalogItems/"+
		"%s", catalogItemID)
	url := client.BuildEncodedURL(path, nil)
	resp, respErr := client.Get(url, nil)
	if respErr != nil {
		return "", respErr
	}

	var response utils.CatalogItem
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &response)
	if unmarshallErr != nil {
		return "", unmarshallErr
	}
	return response.CatalogItem.Name, nil
}

// ReadCatalogItemIdByName - To read id of catalog from vRA using catalog_name
// func ReadCatalogItemByName(catalogName string) (string, error) {
//
// 	path := "/catalog-service/api/consumer/entitledCatalogItemViews"
// 	log.Info("Fetching business group id from name..GET %s ", path)
// 	uri := client.BuildEncodedURL(path+"?$filter=name eq "+catalogName, nil)
// 	customURL := strings.Replace(uri, "%3F", "?", -1)
//
// 	respBody, respErr := client.Get(customURL, nil)
// 	if respErr != nil {
// 		return "", respErr
// 	}
//
// 	var response CatalogItem
// 	unmarshallErr := utils.UnmarshalJSON(respBody, &response)
// 	if unmarshallErr != nil {
// 		return "", unmarshallErr
// 	}
// 	log.Info("the catalog item is %v ", response.CatalogItem.ID)
// 	return response.CatalogItem.ID, nil
// }

//readCatalogItemIdByName - To read id of catalog from vRA using catalog_name
func ReadCatalogItemByName(catalogName string) (string, error) {
	var catalogItemID string

	log.Info("readCatalogItemIdByName->catalog_name %v\n", catalogName)

	//Set a call to read number of catalogs from vRA
	path := fmt.Sprintf("catalog-service/api/consumer/entitledCatalogItemViews")

	url := client.BuildEncodedURL(path, nil)
	resp, respErr := client.Get(url, nil)
	if respErr != nil || resp.StatusCode != 200 {
		return "", respErr
	}

	var template utils.EntitledCatalogItemViews
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &template)
	if unmarshallErr != nil {
		return "", unmarshallErr
	}

	var catalogItemNameArray []string
	interfaceArray := template.Content.([]interface{})
	catalogItemNameLen := len(catalogName)

	//Iterate over all catalog results to find out matching catalog name
	// provided in terraform configuration file
	for i := range interfaceArray {
		catalogItem := interfaceArray[i].(map[string]interface{})
		if catalogItemNameLen <= len(catalogItem["name"].(string)) {
			//If exact name matches then return respective catalog_id
			//else if provided catalog matches as a substring in name then store it in array
			if catalogName == catalogItem["name"].(string) {
				return catalogItem["catalogItemId"].(string), nil
			} else if catalogName == catalogItem["name"].(string)[0:catalogItemNameLen] {
				catalogItemNameArray = append(catalogItemNameArray, catalogItem["name"].(string))
			}
		}
	}

	// If multiple catalog items are present with provided catalog_name
	// then raise an error and show all names of catalog items with similar name
	if len(catalogItemNameArray) > 0 {
		for index := range catalogItemNameArray {
			catalogItemNameArray[index] = strconv.Itoa(index+1) + " " + catalogItemNameArray[index]
		}
		errorMessage := strings.Join(catalogItemNameArray, "\n")
		punctuation := "is"
		if len(catalogItemNameArray) > 1 {
			punctuation = "are"
		}
		return "", fmt.Errorf("There %s total %d catalog(s) present with same name.\n%s\n"+
			"Please select from above.", punctuation, len(catalogItemNameArray), errorMessage)
	}
	return catalogItemID, nil
}

// GetBusinessGroupID retrieves business group id from business group name
func GetBusinessGroupID(businessGroupName string, tenant string) (string, error) {

	path := "/identity/api/tenants/" + tenant + "/subtenants"

	log.Info("Fetching business group id from name..GET %s ", path)

	uri := client.BuildEncodedURL(path+"?$filter=name eq "+businessGroupName, nil)
	customURL := strings.Replace(uri, "%3F", "?", -1)

	//url := client.BuildEncodedURL(cus, nil)
	resp, respErr := client.Get(customURL, nil)
	if respErr != nil {
		return "", respErr
	}

	var businessGroup utils.BusinessGroups
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &businessGroup)
	if unmarshallErr != nil {
		return "", unmarshallErr
	}
	// BusinessGroups array will contain only one BusinessGroup element containing the BG
	// with the name businessGroupName.
	// Fetch the id of that BG
	return businessGroup.Content[0].ID, nil
}

//DestroyMachine - To set resource destroy call
func DestroyMachine(destroyTemplate *utils.ResourceActionTemplate, destroyActionURL string) error {

	buffer, _ := utils.MarshalToJSON(destroyTemplate)
	url := client.BuildEncodedURL(destroyActionURL, nil)
	resp, respErr := client.Post(url, buffer, nil)
	if respErr != nil || resp.StatusCode != 201 {
		return respErr
	}
	return nil
}

//GetRequestStatus - To read request status of resource
// which is used to show information to user post create call.
func GetRequestStatus(requestID string) (*utils.RequestStatusView, error) {
	//Form a URL to read request status
	path := fmt.Sprintf("catalog-service/api/consumer/requests/%s", requestID)
	url := client.BuildEncodedURL(path, nil)
	resp, respErr := client.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var response utils.RequestStatusView
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &response)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &response, nil
}

// GetDeploymentState - Read the state of a vRA7 Deployment
func GetDeploymentState(CatalogRequestID string) (*utils.ResourceView, error) {
	//Form an URL to fetch resource list view
	path := fmt.Sprintf(utils.GetRequestResourceViewAPI, CatalogRequestID)
	url := client.BuildEncodedURL(path, nil)
	resp, respErr := client.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var response utils.ResourceView
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &response)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &response, nil
}

// GetRequestResourceView retrieves the resources that were provisioned as a result of a given request.
func GetRequestResourceView(catalogRequestID string) (*utils.RequestResourceView, error) {
	path := fmt.Sprintf(utils.GetRequestResourceViewAPI, catalogRequestID)
	url := client.BuildEncodedURL(path, nil)
	resp, respErr := client.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var response utils.RequestResourceView
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &response)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &response, nil
}

//RequestCatalogItem - Make a catalog request.
func RequestCatalogItem(requestTemplate *utils.CatalogItemRequestTemplate) (*utils.CatalogRequest, error) {
	//Form a path to set a REST call to create a machine
	path := fmt.Sprintf("/catalog-service/api/consumer/entitledCatalogItems/%s"+
		"/requests", requestTemplate.CatalogItemID)

	buffer, _ := utils.MarshalToJSON(requestTemplate)
	url := client.BuildEncodedURL(path, nil)
	resp, respErr := client.Post(url, buffer, nil)
	if respErr != nil {
		return nil, respErr
	}

	var response utils.CatalogRequest
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &response)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &response, nil
}

// GetResourceActions get the resource actions allowed for a resource
func GetResourceActions(catalogItemRequestID string) (*utils.ResourceActions, error) {
	path := fmt.Sprintf(utils.GetResourceAPI, catalogItemRequestID)

	url := client.BuildEncodedURL(path, nil)
	resp, respErr := client.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var resourceActions utils.ResourceActions
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &resourceActions)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &resourceActions, nil
}

// GetResourceActionTemplate get the action template corresponding to the action id
func GetResourceActionTemplate(resourceID, actionID string) (*utils.ResourceActionTemplate, error) {
	getActionTemplatePath := fmt.Sprintf(utils.GetActionTemplateAPI, resourceID, actionID)
	log.Info("Call GET to fetch the reconfigure action template %v ", getActionTemplatePath)
	url := client.BuildEncodedURL(getActionTemplatePath, nil)
	resp, respErr := client.Get(url, nil)
	if respErr != nil {
		return nil, respErr
	}

	var resourceActionTemplate utils.ResourceActionTemplate
	unmarshallErr := utils.UnmarshalJSON(resp.Body, &resourceActionTemplate)
	if unmarshallErr != nil {
		return nil, unmarshallErr
	}
	return &resourceActionTemplate, nil
}

// PostResourceConfig updates the resource
func PostResourceConfig(reconfigPostLink string, resourceActionTemplate *utils.ResourceActionTemplate) error {

	buffer, _ := utils.MarshalToJSON(resourceActionTemplate)
	url := client.BuildEncodedURL(reconfigPostLink, nil)
	resp, respErr := client.Post(url, buffer, nil)
	if respErr != nil || resp.StatusCode != 201 {
		return respErr
	}
	return nil
}