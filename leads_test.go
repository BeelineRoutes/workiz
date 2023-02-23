
package workiz 

import (
	"github.com/stretchr/testify/assert"

	"testing"
	"encoding/json"
)

func TestLeadResponseStruct (t *testing.T) {
	resp := &leadResponse{}

	err := json.Unmarshal([]byte(`{"flag":true,"data":[{"UUID":"SRUYUI","SerialId":"3","LeadDateTime":"2023-02-28 12:00:00","LeadEndDateTime":"2023-02-28 13:00:00","CreatedDate":"2022-12-19 09:44:07","ClientId":"1002","Status":"new","SubStatus":"","PaymentDueDate":"2023-01-18 00:00:00","Phone":"","SecondPhone":"","PhoneExt":"","SecondPhoneExt":"","Email":"nathan+1@beelineroutes.com","Comments":"","FirstName":"Nathan","LastName":"Thomas","Company":"","Address":"23 Potter pl","City":"Shelburne","State":"VT","PostalCode":"05482","Country":"US","Unit":"","Latitude":"44.3998458","Longitude":"-73.2037722","LeadNotes":"","JobSource":"","CreatedBy":"Nathan Thomas","Team":[],"JobType":"Growler Fill"},{"UUID":"ET38H9","SerialId":"4","LeadDateTime":"2023-02-27 14:00:00","LeadEndDateTime":"2023-02-27 14:15:00","CreatedDate":"2022-12-19 09:44:42","ClientId":"1002","Status":"new","SubStatus":"","PaymentDueDate":"2023-01-18 00:00:00","Phone":"","SecondPhone":"","PhoneExt":"","SecondPhoneExt":"","Email":"nathan+1@beelineroutes.com","Comments":"","FirstName":"Nathan","LastName":"Thomas","Company":"","Address":"23 Potter pl","City":"Shelburne","State":"VT","PostalCode":"05482","Country":"US","Unit":"","Latitude":"44.3998458","Longitude":"-73.2037722","LeadNotes":"","JobSource":"","CreatedBy":"Nathan Thomas","Team":[{"id":"228777","name":"Nathan Thomas"},{"id":"246389","name":"Brooklyn Thomas"}],"JobType":"Full Case"},{"UUID":"Z7X968","SerialId":"2","LeadDateTime":"2022-10-02 15:00:00","LeadEndDateTime":"2022-10-02 16:00:00","CreatedDate":"2022-09-07 12:47:20","ClientId":"1002","Status":"new","SubStatus":"","PaymentDueDate":"2022-10-07 00:00:00","Phone":"","SecondPhone":"","PhoneExt":"","SecondPhoneExt":"","Email":"nathan+1@beelineroutes.com","Comments":"","FirstName":"Nathan","LastName":"Thomas","Company":"","Address":"23 Potter pl","City":"Shelburne","State":"VT","PostalCode":"05482","Country":"US","Unit":"","Latitude":"44.3998458","Longitude":"-73.2037722","LeadNotes":"","JobSource":"","CreatedBy":"Nathan Thomas","Team":[],"JobType":"Growler Fill"},{"UUID":"9T3O0W","SerialId":"1","LeadDateTime":"2022-10-01 15:00:00","LeadEndDateTime":"2022-10-01 16:00:00","CreatedDate":"2022-09-07 12:44:36","ClientId":"1002","Status":"new","SubStatus":"","PaymentDueDate":"2022-10-07 00:00:00","Phone":"","SecondPhone":"","PhoneExt":"","SecondPhoneExt":"","Email":"nathan+1@beelineroutes.com","Comments":"","FirstName":"","LastName":"","Company":"","Address":"","City":"","State":"","PostalCode":"","Country":"US","Unit":"","Latitude":"","Longitude":"","LeadNotes":"","JobSource":"","CreatedBy":"Nathan Thomas","Team":[],"JobType":"Growler Fill"}],"has_more":false,"found":4,"code":200}`), resp)

	if err != nil { t.Fatal(err) }
	assert.Equal (t, 4, len(resp.Data))

}

