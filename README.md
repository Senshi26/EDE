# Emarsys Duplicates Exterminator

The idea came along years of struggle with records' duplications within Emarsys platform. 
<br>This repository is dedicated to our precious customers.<br>
Use it to keep your database clean :relieved:


Installation: 

`go get github.com/Senshi26/EDE/`

Usage: 



Using Merge Rules \/

```go
package main

import (
	ede "emarsys_duplicates_exterminator"
	"fmt"
)

func main() {
 config :=ede.EdeData{
	Emarsys_auth: ede.SuiteAPI{User:"XXXXXXXXX",Secret:"XXXXXXXXXXXXXXXXX"}, //API creds required to authenticate
	SearchField:  "3",//Sets the field_id which will be used as unique key to search duplications
	MergeRules:   ede.MergeRules{ByDateField:"3842",//Sorts the duplications based on date field_id you specify
		UpdateEmptyField:true, // Will populate empty fields in primary record with the latest duplicate's record values
		CreateContactList:true},//Adds a contactlist into platform with  processed primary records
    /*	LastAdded:true} */ //Will use the last added record as primary
	}


	err := config.FindDuplicates("duplicate_by_email@gmail.com") //receives the value of unique key to search

	if err != nil{

		fmt.Println(err)

	}

}


```

Using Exclusion when field 530 is empty \/

```go
package main

import (
	ede "emarsys_duplicates_exterminator"
	"fmt"
)

func main() {
	config := ede.EdeData{
		Emarsys_auth: ede.SuiteAPI{username, secret}, //API creds required to authenticate
		SearchField:  "3",                            //Sets the field_id which will be used as  key to search duplications
		Exclude: ede.Exclude{
			FieldId:    "530", // Exclude contacts from deletion/merging with following field_id in Emarsys 
			FieldValue: ede.Field_value{Null: true}, // rule will be applied if value of field is empty for true and not empty for false
		},

	}


	err := config.FindDuplicates("duplicate_by_email@gmail.com") //receives the value of unique key to search

	if err != nil{

		fmt.Println(err)

	}

}


```



 #### Known limitations: ####
EDE will not merge > 1000 duplicates by unique key in a single query

:8ball:


