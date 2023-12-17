/*
 * Copyright 2023 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// [START setup]
// [START imports]
package main

import (
	"bytes"
        "context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauthJwt "golang.org/x/oauth2/jwt"
        "google.golang.org/api/option"
        "google.golang.org/api/walletobjects/v1"
	"io"
	"os"
	"strings"
)

// [END imports]

const (
	batchUrl  = "https://walletobjects.googleapis.com/batch"
	classUrl  = "https://walletobjects.googleapis.com/walletobjects/v1/loyaltyClass"
	objectUrl = "https://walletobjects.googleapis.com/walletobjects/v1/loyaltyObject"
)

// [END setup]

type demoLoyalty struct {
	credentials                   *oauthJwt.Config
	service                       *walletobjects.Service
	batchUrl, classUrl, objectUrl string
}

// [START auth]
// Create authenticated HTTP client using a service account file.
func (d *demoLoyalty) auth() {
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	b, _ := os.ReadFile(credentialsFile)
	credentials, err := google.JWTConfigFromJSON(b, walletobjects.WalletObjectIssuerScope)
	if err != nil {
		log.Fatalf("Unable to create credentials: %v", err)
	}
	d.credentials = credentials
	d.service, _ = walletobjects.NewService(context.Background(), option.WithCredentialsFile(credentialsFile))
}

// [END auth]

// [START createClass]
// Create a class.
func (d *demoLoyalty) createClass(issuerId, classSuffix string) {
	logo := walletobjects.Image {
		SourceUri: &walletobjects.ImageUri {
			Uri: "http://farm8.staticflickr.com/7340/11177041185_a61a7f2139_o.jpg",
		},
	}
	loyaltyClass := new(walletobjects.LoyaltyClass)
	loyaltyClass.Id = fmt.Sprintf("%s.%s", issuerId, classSuffix)
	loyaltyClass.ProgramName = "Program name"
	loyaltyClass.IssuerName = "Issuer name"
	loyaltyClass.ReviewStatus = "UNDER_REVIEW"
	loyaltyClass.ProgramLogo = &logo
	res, err := d.service.Loyaltyclass.Insert(loyaltyClass).Do()
	if err != nil {
		log.Fatalf("Unable to insert class: %v", err)
	} else {
		fmt.Printf("Class insert id:\n%v\n", res.Id)
	}
}

// [END createClass]

// [START createObject]
// Create an object.
func (d *demoLoyalty) createObject(issuerId, classSuffix, objectSuffix string) {
	loyaltyObject := new(walletobjects.LoyaltyObject)
	loyaltyObject.Id = fmt.Sprintf("%s.%s", issuerId, objectSuffix)
	loyaltyObject.ClassId = fmt.Sprintf("%s.%s", issuerId, classSuffix)
	loyaltyObject.AccountName = "Account name"
	loyaltyObject.AccountId = "Account id"
	loyaltyObject.State = "ACTIVE"
	loyaltyObject.LoyaltyPoints = &walletobjects.LoyaltyPoints {
		Balance: &walletobjects.LoyaltyPointsBalance { Int: 800 },
		Label: "Points",
	}
	loyaltyObject.HeroImage = &walletobjects.Image {
		SourceUri: &walletobjects.ImageUri {
			Uri: "https://farm4.staticflickr.com/3723/11177041115_6e6a3b6f49_o.jpg",
		},
	}
	loyaltyObject.Barcode = &walletobjects.Barcode {
		Type: "QR_CODE",
		Value: "QR code",
	}
	loyaltyObject.Locations = []*walletobjects.LatLongPoint {
		&walletobjects.LatLongPoint {
			Latitude: 37.424015499999996,
			Longitude: -122.09259560000001,
		},
	}
	loyaltyObject.LinksModuleData = &walletobjects.LinksModuleData {
		Uris: []*walletobjects.Uri {
			&walletobjects.Uri {
				Id: "LINK_MODULE_URI_ID",
				Uri: "http://maps.google.com/",
				Description: "Link module URI description",
			},
			&walletobjects.Uri {
				Id: "LINK_MODULE_TEL_ID",
				Uri: "tel:6505555555",
				Description: "Link module tel description",
			},
		},
	}
	loyaltyObject.ImageModulesData = []*walletobjects.ImageModuleData {
		&walletobjects.ImageModuleData {
			Id: "IMAGE_MODULE_ID",
			MainImage: &walletobjects.Image {
				SourceUri: &walletobjects.ImageUri{
					Uri: "http://farm4.staticflickr.com/3738/12440799783_3dc3c20606_b.jpg",
				},
			},
		},
	}
	loyaltyObject.TextModulesData = []*walletobjects.TextModuleData {
		&walletobjects.TextModuleData {
			Body: "Text module body",
			Header: "Text module header",
			Id: "TEXT_MODULE_ID",
		},
	}
	res, err := d.service.Loyaltyobject.Insert(loyaltyObject).Do()
	if err != nil {
		log.Fatalf("Unable to insert object: %v", err)
	} else {
		fmt.Printf("Object insert id:\n%s\n", res.Id)
	}
}

// [END createObject]

// [START expireObject]
// Expire an object.
//
// Sets the object's state to Expired. If the valid time interval is
// already set, the pass will expire automatically up to 24 hours after.
func (d *demoLoyalty) expireObject(issuerId, objectSuffix string) {
	loyaltyObject := &walletobjects.LoyaltyObject {
		State: "EXPIRED",
	}
	res, err := d.service.Loyaltyobject.Patch(fmt.Sprintf("%s.%s", issuerId, objectSuffix), loyaltyObject).Do()
	if err != nil {
		log.Fatalf("Unable to patch object: %v", err)
	} else {
		fmt.Printf("Object expiration id:\n%s\n", res.Id)
	}
}

// [END expireObject]

// [START jwtNew]
// Generate a signed JWT that creates a new pass class and object.
//
// When the user opens the "Add to Google Wallet" URL and saves the pass to
// their wallet, the pass class and object defined in the JWT are
// created. This allows you to create multiple pass classes and objects in
// one API call when the user saves the pass to their wallet.
func (d *demoLoyalty) createJwtNewObjects(issuerId, classSuffix, objectSuffix string) {
	loyaltyObject := new(walletobjects.LoyaltyObject)
	loyaltyObject.Id = fmt.Sprintf("%s.%s", issuerId, objectSuffix + "_")
	loyaltyObject.ClassId = fmt.Sprintf("%s.%s", issuerId, classSuffix)
	loyaltyObject.AccountName = "Account name"
	loyaltyObject.AccountId = "Account id"
	loyaltyObject.State = "ACTIVE"
	loyaltyJson, _ := json.Marshal(loyaltyObject)
	var payload map[string]any
	json.Unmarshal([]byte(fmt.Sprintf(`
	{
		"loyaltyObjects": [%s]
	}
	`, loyaltyJson)), &payload)
	claims := jwt.MapClaims{
		"iss":     d.credentials.Email,
		"aud":     "google",
		"origins": []string{"www.example.com"},
		"typ":     "savetowallet",
		"payload": payload,
	}

	// The service account credentials are used to sign the JWT
	key, _ := jwt.ParseRSAPrivateKeyFromPEM(d.credentials.PrivateKey)
	token, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)

	fmt.Println("Add to Google Wallet link for new objects")
	fmt.Println("https://pay.google.com/gp/v/save/" + token)
}

// [END jwtNew]

// [START jwtExisting]
// Generate a signed JWT that references an existing pass object.

// When the user opens the "Add to Google Wallet" URL and saves the pass to
// their wallet, the pass objects defined in the JWT are added to the
// user's Google Wallet app. This allows the user to save multiple pass
// objects in one API call.
func (d *demoLoyalty) createJwtExistingObjects(issuerId string, classSuffix string, objectSuffix string) {
	loyaltyObject := new(walletobjects.LoyaltyObject)
	loyaltyObject.Id = fmt.Sprintf("%s.%s", issuerId, objectSuffix)
	loyaltyObject.ClassId = fmt.Sprintf("%s.%s", issuerId, classSuffix)
	loyaltyJson, _ := json.Marshal(loyaltyObject)
	var payload map[string]any
	json.Unmarshal([]byte(fmt.Sprintf(`
	{
		"loyaltyObjects":[%s]
	}
	`, loyaltyJson )), &payload)
	claims := jwt.MapClaims{
		"iss":     d.credentials.Email,
		"aud":     "google",
		"origins": []string{"www.example.com"},
		"typ":     "savetowallet",
		"payload": payload,
	}

	// The service account credentials are used to sign the JWT
	key, _ := jwt.ParseRSAPrivateKeyFromPEM(d.credentials.PrivateKey)
	token, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)

	fmt.Println("Add to Google Wallet link for existing objects")
	fmt.Println("https://pay.google.com/gp/v/save/" + token)
}

// [END jwtExisting]

// [START batch]
// Batch create Google Wallet objects from an existing class.
func (d *demoLoyalty) batchCreateObjects(issuerId, classSuffix string) {
	data := ""
	for i := 0; i < 3; i++ {
		objectSuffix := strings.ReplaceAll(uuid.New().String(), "-", "_")

		batchObject := fmt.Sprintf(`
		{
			"classId": "%s.%s",
			"heroImage": {
				"contentDescription": {
					"defaultValue": {
						"value": "Hero image description",
						"language": "en-US"
					}
				},
				"sourceUri": {
					"uri": "https://farm4.staticflickr.com/3723/11177041115_6e6a3b6f49_o.jpg"
				}
			},
			"barcode": {
				"type": "QR_CODE",
				"value": "QR code"
			},
			"locations": [
				{
					"latitude": 37.424015499999996,
					"longitude": -122.09259560000001
				}
			],
			"accountName": "Account name",
			"state": "ACTIVE",
			"linksModuleData": {
				"uris": [
					{
						"id": "LINK_MODULE_URI_ID",
						"uri": "http://maps.google.com/",
						"description": "Link module URI description"
					},
					{
						"id": "LINK_MODULE_TEL_ID",
						"uri": "tel:6505555555",
						"description": "Link module tel description"
					}
				]
			},
			"imageModulesData": [
				{
					"id": "IMAGE_MODULE_ID",
					"mainImage": {
						"contentDescription": {
							"defaultValue": {
								"value": "Image module description",
								"language": "en-US"
							}
						},
						"sourceUri": {
							"uri": "http://farm4.staticflickr.com/3738/12440799783_3dc3c20606_b.jpg"
						}
					}
				}
			],
			"textModulesData": [
				{
					"body": "Text module body",
					"header": "Text module header",
					"id": "TEXT_MODULE_ID"
				}
			],
			"id": "%s.%s",
			"loyaltyPoints": {
				"balance": {
					"int": 800
				},
				"label": "Points"
			},
			"accountId": "Account id"
		}
		`, issuerId, classSuffix, issuerId, objectSuffix)

		data += "--batch_createobjectbatch\n"
		data += "Content-Type: application/json\n\n"
		data += "POST /walletobjects/v1/loyaltyObject\n\n"
		data += batchObject + "\n\n"
	}
	data += "--batch_createobjectbatch--"

	res, err := d.credentials.Client(oauth2.NoContext).Post(batchUrl, "multipart/mixed; boundary=batch_createobjectbatch", bytes.NewBuffer([]byte(data)))

	if err != nil {
		fmt.Println(err)
	} else {
		b, _ := io.ReadAll(res.Body)
		fmt.Printf("Batch insert response:\n%s\n", b)
	}
}

// [END batch]

func main() {
	if len(os.Args) == 0 {
		fmt.Println("Usage: go run demo_loyalty.go <issuer-id>")
		os.Exit(1)
	}

	issuerId := os.Getenv("WALLET_ISSUER_ID")
	classSuffix := strings.ReplaceAll(uuid.New().String(), "-", "_")
	objectSuffix := fmt.Sprintf("%s-%s", strings.ReplaceAll(uuid.New().String(), "-", "_"), classSuffix)

	d := demoLoyalty{}

	d.auth()
	d.createClass(issuerId, classSuffix)
	d.createObject(issuerId, classSuffix, objectSuffix)
	d.expireObject(issuerId, objectSuffix)
	d.createJwtNewObjects(issuerId, classSuffix, objectSuffix)
	d.createJwtExistingObjects(issuerId, classSuffix, objectSuffix)
	d.batchCreateObjects(issuerId, classSuffix)
}
