package ede

import "time"

type SuiteAPI struct {
	User   string
	Secret string
}

type EdeData struct {
	Emarsys_auth SuiteAPI
	SearchField  string
	MergeRules   MergeRules
	Exclude      Exclude
}

type MergeRules struct {
	LastAdded         bool
	ByDateField       string
	UpdateEmptyField  bool
	CreateContactList bool
}

type Exclude struct {
	FieldId    string
	FieldValue Field_value
}

type Field_value struct {
	Null bool
}

type DataQueryResponse struct {
	Data struct {
		Errors []interface{} `json:"errors"`
		Result []interface{} `json:"result"`
	} `json:"data"`
	ReplyCode int64  `json:"replyCode"`
	ReplyText string `json:"replyText"`
}

type Settings struct {
	ReplyCode int    `json:"replyCode"`
	ReplyText string `json:"replyText"`
	Data      struct {
		ID                       int    `json:"id"`
		Environment              string `json:"environment"`
		Timezone                 string `json:"timezone"`
		Name                     string `json:"name"`
		PasswordHistoryQueueSize int    `json:"password_history_queue_size"`
		Country                  string `json:"country"`
		TotalContacts            string `json:"totalContacts"`
	} `json:"data"`
}

type ReturnedDupsList struct {
	ReplyCode int    `json:"replyCode"`
	ReplyText string `json:"replyText"`
	Data      struct {
		Errors []interface{} `json:"errors"`
		Result []struct {
			Optin string `json:"31"`
			ID    string `json:"id"`
		} `json:"result"`
	} `json:"data"`
}

type DupsByDate struct {
	ReplyCode int    `json:"replyCode"`
	ReplyText string `json:"replyText"`
	Data      struct {
		Errors []interface{} `json:"errors"`
		Result []struct {
			DateField string `json:"date_field"`
			ID        string `json:"id"`
		} `json:"result"`
	} `json:"data"`
}

type Date_dups_slice []struct {
	ID         string
	Date_field time.Time
}

type Date_dups_slice_element struct {
	ID         string
	Date_field time.Time
}

type EmarsysFields struct {
	ReplyCode int    `json:"replyCode"`
	ReplyText string `json:"replyText"`
	Data      []struct {
		ID              int    `json:"id"`
		Name            string `json:"name"`
		ApplicationType string `json:"application_type"`
		StringID        string `json:"string_id"`
	} `json:"data"`
}

type Suite_Contact_Response struct {
	ReplyCode int    `json:"replyCode"`
	ReplyText string `json:"replyText"`
	Data      struct {
		Errors []interface{} `json:"errors"`
		Result []interface{} `json:"result"`
	} `json:"data"`
}
type Cl_response struct {
	Data struct {
		Errors []interface{} `json:"errors,omitempty"`
		ID     int           `json:"id,omitempty"`
	} `json:"data,omitempty"`
	ReplyCode int    `json:"replyCode"`
	ReplyText string `json:"replyText"`
}

type CL_List struct {
	Data []struct {
		Created string `json:"created,omitempty"`
		ID      string `json:"id,omitempty"`
		Name    string `json:"name,omitempty"`
		Type    int    `json:"type,omitempty"`
	} `json:"data"`
	ReplyCode int    `json:"replyCode"`
	ReplyText string `json:"replyText"`
}
