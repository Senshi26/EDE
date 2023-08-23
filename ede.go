package ede

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var emarsysFields EmarsysFields

type EDE interface {
	FindDuplicates(searchValue string) error
	FindDuplicatesExclude(searchValue string) error
}

func (EData EdeData) FindDuplicatesExclude(searchValue string) error {

	checkAuth, err := EData.Emarsys_auth.CheckAuth()

	if err != nil {

		return err

	} else if checkAuth {

		if EData.Exclude.FieldId != "" {

			dups_list, err := EData.GetByLastAdded(searchValue)

			if err != nil {

				fmt.Println(err)
				return err
			}
			str_start := `{
  "keyId": "id",
  "keyValues": [
    `

			str_mid := ""

			for i := range dups_list {

				str_mid += `"` + strconv.Itoa(dups_list[i]) + "\","

			}

			str_end := `],
  "fields": [
    "3","` + EData.Exclude.FieldId + `"
  ]
}`

			full_str := str_start + str_mid[:len(str_mid)-1] + str_end

			fmt.Println(full_str)

			_, get_emails := EData.Emarsys_auth.send("POST", "contact/getdata", full_str)

			var get_Data DataQueryResponse

			get_emails = JSON_FIX(get_emails)

			json.Unmarshal([]byte(get_emails), &get_Data)

			for j := range get_Data.Data.Result {

				_, ok := get_Data.Data.Result[j].(map[string]interface{})
				if ok && len(get_Data.Data.Result) > 0 {

					fmt.Println("Значение из полученного списка по getData",get_Data.Data.Result[j].(map[string]interface{})[EData.Exclude.FieldId])


					if EData.Exclude.FieldId != "" && !EData.Exclude.FieldValue.Null &&
						get_Data.Data.Result[j].(map[string]interface{})[EData.Exclude.FieldId] != "" {

						continue

					} else if EData.Exclude.FieldId != "" && EData.Exclude.FieldValue.Null &&
						get_Data.Data.Result[j].(map[string]interface{})[EData.Exclude.FieldId] == "" {

						continue

					} else {

						EData.Emarsys_auth.send("POST", "contact/delete", `{ "key_id": "id", "id": "`+get_Data.Data.Result[j].(map[string]interface{})["id"].(string) +`" }`)

					}

				}

			}
		}

	}
	return nil
}

func (EData EdeData) FindDuplicates(searchValue string) error {

	checkAuth, err := EData.Emarsys_auth.CheckAuth()

	if err != nil {

		return err

	} else if checkAuth {

		switch EData.SearchField {

		case "":

			return errors.New("search field must not be empty")

		}

		if EData.MergeRules.LastAdded {
			if EData.MergeRules.UpdateEmptyField {

				dups_slice, err := EData.GetByLastAdded(searchValue)

				if err != nil {

					fmt.Println(err)

				}

				var dupsSliceStr string

				missing_fields := make(map[string]string)

				if len(dups_slice) > 0 {

					for h := range dups_slice {

						dupsSliceStr += `"` + strconv.Itoa(dups_slice[h]) + `",`

					}

					dataRequest := `{
										  "keyId": "id",
										  "keyValues": [` +
						dupsSliceStr[0:len(dupsSliceStr)-1] +
						`]
										}`

					_, result := EData.Emarsys_auth.send("POST", "contact/getdata", dataRequest)

					bf := bytes.NewBuffer([]byte{})
					jsonEncoder := json.NewEncoder(bf)
					jsonEncoder.SetEscapeHTML(false)
					jsonEncoder.Encode(result)

					result = JSON_FIX(result)

					err3 := EData.GetEmarsysFields()

					if err3 != nil {

						fmt.Println(err)

					}

					missing_fields, err = CompareFields(result, emarsysFields)

					if err != nil {

						fmt.Println(err)

					}

				} else {

					return errors.New("No duplicates found")

				}

				if len(emarsysFields.Data) > 0 {

					for l := 1; l <= len(dups_slice)-1; l++ {

						EData.Emarsys_auth.send("POST", "contact/delete", `{ "key_id": "id", "id": "`+strconv.Itoa(dups_slice[l])+`" }`)

					}

					switch EData.MergeRules.CreateContactList {

					case true:
						err := EData.CreateContactList(searchValue, "Merged duplicates "+time.Now().Format("2006-01-02"))

						if err != nil {

							panic(err)
						}

						err2 := EData.UpdateContactMissingFields(missing_fields, strconv.Itoa(dups_slice[0]))

						if err2 != nil {

							fmt.Println(err2)

						}

					case false:

						err3 := EData.UpdateContactMissingFields(missing_fields, strconv.Itoa(dups_slice[0]))

						if err3 != nil {

							fmt.Println(err3)

						}
					}

				} else {

					return errors.New("list of fields is empty")

				}

			} else {
				dups_slice, err := EData.GetByLastAdded(searchValue)

				if err != nil {

					panic(err)

				}

				for l := 1; l <= len(dups_slice)-1; l++ {

					EData.Emarsys_auth.send("POST", "contact/delete", `{ "key_id": "id", "id": "`+strconv.Itoa(dups_slice[l])+`" }`)

				}
			}
		} else if EData.MergeRules.ByDateField != "" && EData.MergeRules.LastAdded {

			return errors.New("logical error:\n two mutually exclusive conditions")

		} else if EData.MergeRules.ByDateField != "" {

			_, err := strconv.Atoi(EData.MergeRules.ByDateField)

			if err != nil {

				return errors.New("provided date field_id is not an integer: " + EData.MergeRules.ByDateField)

			}

			if EData.MergeRules.UpdateEmptyField {

				dateDupsSlice, err := EData.GetByDateField(searchValue)

				if err != nil {

					fmt.Println(err)

				}

				var dupsSliceStr string

				for h := range dateDupsSlice {

					dupsSliceStr += `"` + dateDupsSlice[h].ID + `",`

				}

				dataRequest := `{
										  "keyId": "id",
										  "keyValues": [` +
					dupsSliceStr[0:len(dupsSliceStr)-1] +
					`]
										}`

				_, result := EData.Emarsys_auth.send("POST", "contact/getdata", dataRequest)

				result = JSON_FIX(result)

				err4 := EData.GetEmarsysFields()

				if err4 != nil {

					panic(err)

				}

				if len(emarsysFields.Data) > 0 {

					missing_fields, err := CompareFields(result, emarsysFields)

					if err != nil {

						fmt.Println(err)

					}

					err6 := EData.UpdateContactMissingFields(missing_fields, dateDupsSlice[0].ID)

					if err6 != nil {

						fmt.Println(err)

					}

					for l := 1; l <= len(dateDupsSlice)-1; l++ {

						EData.Emarsys_auth.send("POST", "contact/delete", `{ "key_id": "id", "id": "`+dateDupsSlice[l].ID+`" }`)

					}

					switch EData.MergeRules.CreateContactList {

					case true:

						fieldName, err := GetFieldName(emarsysFields, EData.MergeRules.ByDateField)

						err7 := EData.CreateContactList(dateDupsSlice[0].ID, "Duplicates by date field "+fieldName+" "+time.Now().Format("2006-01-02"))

						if err != nil {

							panic(err7)

						}

					}

				} else {

					return errors.New("list of fields is empty")

				}

			} else {

				dateDupsSlice, err := EData.GetByDateField(searchValue)

				if err != nil {

					panic(err)

				}

				for l := 1; l <= len(dateDupsSlice)-1; l++ {

					EData.Emarsys_auth.send("POST", "contact/delete", `{ "key_id": "id", "id": "`+dateDupsSlice[l].ID+`" }`)

				}

			}

		}

	}

	return nil
}
