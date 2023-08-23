package ede

import (
	"bytes"
	"crypto/sha1"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

// Sends an HTTP request to the Emarsys API
func (config SuiteAPI) send(method string, path string, body string) (string, string) {
	url := "https://api.emarsys.net/api/v2/" + path
	var timestamp = time.Now().Format(time.RFC3339)
	nonce := generateRandString(36)
	text := (nonce + timestamp + config.Secret)
	h := sha1.New()
	h.Write([]byte(text))
	sha1 := hex.EncodeToString(h.Sum(nil))
	passwordDigest := b64.StdEncoding.EncodeToString([]byte(sha1))

	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	header := string(" UsernameToken Username=\"" + config.User + "\",PasswordDigest=\"" + passwordDigest + "\",Nonce=\"" + nonce + "\",Created=\"" + timestamp + "\"")

	req.Header.Set("X-WSSE", header)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	status := resp.Status
	responseBody, _ := ioutil.ReadAll(resp.Body)
	return status, string(responseBody)
}

func CompareFields(response string, fields EmarsysFields) (map[string]string, error) {

	missing_fields := make(map[string]string)

	end_result := strings.Replace(response, "null", `""`, -1)

	end_result = strings.Replace(end_result, "Null", `""`, -1)

	contact_data := Suite_Contact_Response{}

	err := json.Unmarshal([]byte(end_result), &contact_data)

	if err != nil {

		fmt.Println(err)
		return nil, errors.New("Empty data")

	}

	if len(contact_data.Data.Result) > 1 {

		for y := len(contact_data.Data.Result) - 1; y > 0; y-- {

			_, ok := contact_data.Data.Result[y].(map[string]interface{})
			if ok {

				for u := range fields.Data {
					_, ok2 := contact_data.Data.Result[y].(map[string]interface{})[strconv.Itoa(fields.Data[u].ID)].(string)
					if ok2 {

						if contact_data.Data.Result[0].(map[string]interface{})[strconv.Itoa(fields.Data[u].ID)] == "" &&
							contact_data.Data.Result[y].(map[string]interface{})[strconv.Itoa(fields.Data[u].ID)] != "" &&
							fields.Data[u].ApplicationType != "special" && fields.Data[u].ApplicationType != "voucher" {

							missing_fields[strconv.Itoa(fields.Data[u].ID)] = contact_data.Data.Result[y].(map[string]interface{})[strconv.Itoa(fields.Data[u].ID)].(string)

						}

					}
				}

			}

		}

		return missing_fields, nil

	} else if len(contact_data.Data.Result) == 1 {

		fmt.Println(errors.New("no duplicates found to merge -  single contact found"))

	} else {

		fmt.Println(errors.New("Empty data"))

	}

	return nil, errors.New("Empty data")
}

func (EData EdeData) CreateContactList(contact_id string, list_name string) error {

	_, cl_creation_req := EData.Emarsys_auth.send("POST", "contactlist",
		`{
							  "key_id": "id",
							  "name": "`+list_name+`",
							  "description": "`+`Erased duplicates list_`+time.Now().Format("2006-01-02 15:04:05")+`",
							  "external_ids": [
							"`+contact_id+`"
							  ]
							}`)

	fmt.Println(cl_creation_req)

	cl_creation_req = strings.Replace(cl_creation_req, `,"data":""`, "", -1)

	var cl_response Cl_response

	err := json.Unmarshal([]byte(cl_creation_req), &cl_response)
	if err != nil {

		fmt.Println(err)

	}

	if cl_response.ReplyCode != 0 {

		fmt.Println(cl_response.Data.Errors)
		return errors.New("ContactList has not been created\n")

	}

	return nil

}

func (EData EdeData) GetByLastAdded(searchValue string) ([]int, error) {

	dups_json := ReturnedDupsList{}

	_, returnedDups := EData.Emarsys_auth.send("GET", "contact/query/?"+"return=31"+"&"+EData.SearchField+"="+searchValue, "")

	returnedDups = JSON_FIX(returnedDups)

	err := json.Unmarshal([]byte(returnedDups), &dups_json)

	if err != nil {
		fmt.Println("result: " + returnedDups)
		fmt.Println("url: " + "contact/query/?" + "return=31" + "&" + EData.SearchField + "=" + searchValue)
		return []int{}, err

	}

	var dups_slice []int

	for k := range dups_json.Data.Result {

		contact_id, err := strconv.Atoi(dups_json.Data.Result[k].ID)

		if err != nil {

			return []int{}, errors.New("non integer contact_id found \n" + "please report to TCS\n" + "contact_id: " + dups_json.Data.Result[k].ID)

		}

		//dups_slice += `"` + dups_json.Data.Result[k].ID + `"` + ","
		dups_slice = append(dups_slice, contact_id)

	}

	sort.Slice(dups_slice, func(i, j int) bool {
		return dups_slice[j] < dups_slice[i]
	})

	return dups_slice, nil

}

func (EData EdeData) GetByDateField(searchValue string) (Date_dups_slice, error) {

	_, returnedDups := EData.Emarsys_auth.send("GET", "contact/query/?"+EData.SearchField+"="+searchValue+"&"+"return="+EData.MergeRules.ByDateField, "")

	returnedDups = strings.Replace(returnedDups, `"`+EData.MergeRules.ByDateField+`":`, `"date_field":`, -1)

	dupsByDate := DupsByDate{}

	err := json.Unmarshal([]byte(returnedDups), &dupsByDate)

	if err != nil {

		panic(err)

	}

	dateDupsSlice := Date_dups_slice{}

	for j := range dupsByDate.Data.Result {

		date_field, err := time.Parse("2006-01-02", dupsByDate.Data.Result[j].DateField)

		if err != nil {

			return Date_dups_slice{}, errors.New("duplicate with contact_id = " + dupsByDate.Data.Result[j].ID + " has empty field instead of date field")

		}

		dateDupsSlice = append(dateDupsSlice, Date_dups_slice_element{ID: dupsByDate.Data.Result[j].ID, Date_field: date_field})

	}

	sort.Slice(dateDupsSlice, func(i, j int) bool {
		return dateDupsSlice[i].Date_field.After(dateDupsSlice[j].Date_field)

	})

	return dateDupsSlice, nil

}

func (EData EdeData) UpdateContactMissingFields(missing_fields map[string]string, main_contact string) error {

	var updateStr string

	for k, v := range missing_fields {

		updateStr += `"` + k + `"` + ":" + `"` + v + `",`

	}
	if len(updateStr) > 0 {
		updateStr = `{"key_id": "id","contacts":[{"id":` + `"` + main_contact + `",` + updateStr[0:len(updateStr)-1] + `}]}`
		fmt.Println(updateStr)
		statusCode, response := EData.Emarsys_auth.send("PUT", "contact", updateStr)

		if statusCode != "200" {

			return errors.New(response)

		}

		return nil
	} else {

		return errors.New("Empty data")

	}

}

func (EData EdeData) GetEmarsysFields() error {

	//returns fields list into global variable emarsysFields

	_, temp_fields_string := EData.Emarsys_auth.send("GET", "field", "")

	err2 := json.Unmarshal([]byte(temp_fields_string), &emarsysFields)

	if err2 != nil {

		return err2

	}

	return nil
}

func JSON_FIX(json string) string {

	json = strings.Replace(json, "null", `""`, -1)
	json = strings.Replace(json, "Null", `""`, -1)
	json = strings.Replace(json, "<", `""`, -1)
	json = strings.Replace(json, ">", `""`, -1)
	return json

}

func GetFieldName(emarsysFields EmarsysFields, fieldId string) (string, error) {

	for i := range emarsysFields.Data {

		if fieldId == strconv.Itoa(emarsysFields.Data[i].ID) {

			return emarsysFields.Data[i].Name, nil

		}

	}

	return "", errors.New("Field ID not found")

}

func (Auth SuiteAPI) CheckAuth() (bool, error) {

	if Auth.Secret == "" {

		a := errors.New("Secret is empty")

		return false, a

	} else if Auth.User == "" {

		a := errors.New("User is empty")

		return false, a

	} else {

		json_result := Settings{}

		_, result := Auth.send("GET", "settings", "")

		err := json.Unmarshal([]byte(result), &json_result)

		if err != nil {

			return false, err

		} else {

			return true, nil

		}

	}

}
