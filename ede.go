package ede

import (
	"errors"
	"fmt"

	"strconv"
	"time"
)

var emarsysFields EmarsysFields

type EDE interface {
	FindDuplicates(searchValue string) error
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

					panic(err)

				}

				var dupsSliceStr string

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

				result = JSON_FIX(result)

				err3 := EData.GetEmarsysFields()

				if err3 != nil {

					fmt.Println(err)

				}

				missing_fields, err := CompareFields(result, emarsysFields)

				if err != nil {

					panic(err)

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

							panic(err3)

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

					panic(err)

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

						panic(err)

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
