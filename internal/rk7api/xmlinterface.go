package rk7api

//Logger Debug/Info OK

import (
	"WooWithRkeeper/pkg/logging"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const LSAPI_UCS_URL = "https://l.ucs.ru/ls5api/api/License/"

type xmlInterface struct {
	Usr                      string
	Anchor                   string
	Guid                     string
	SeqNumber                int
	ProductID                string
	RestCode                 int
	ExpirationDateTimeFormat time.Time
	ResultGetLicenseIdByAnchor
}

type ResultGetLicenseIdByAnchor struct {
	Id             string  `json:"id"` //LicenseToken
	ExpirationDate string  `json:"expirationDate"`
	Qty            float64 `json:"qty"`
	ErrorCode      int     `json:"ErrorCode"`
	ErrorMessage   string  `json:"ErrorMessage"`
}

var xmlI *xmlInterface

// получении ID лицензии
func (x *xmlInterface) GetLicenseIdByAnchor() error {
	logger := logging.GetLogger()
	logger.Println("Start GetLicenseIdByAnchor")
	defer logger.Println("End GetLicenseIdByAnchor")

	if x.Id == "" || time.Now().Sub(x.ExpirationDateTimeFormat) >= 0 {
		url := fmt.Sprintf("%s/%s", LSAPI_UCS_URL, "GetLicenseIdByAnchor")
		logger.Debugf("Request:\n%s", url)

		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return errors.Wrap(err, "ошибка при выполнении httprouter.NewRequest")
		}

		x.Anchor = fmt.Sprintf("6:%s#%d/17", x.ProductID, x.RestCode)
		logger.Debugf("Anchor:%s", x.Anchor)

		params := req.URL.Query()
		params.Add("anchor", x.Anchor)

		req.Header.Add("usr", x.Usr)

		req.URL.RawQuery = params.Encode()
		logger.Debugf("RawQuery: %s", req.URL.RawQuery)

		resp, err := client.Do(req)
		if err != nil {
			return errors.Wrap(err, "ошибка при выполнении client.Do")
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				logger.Error(err) //return надо сделать в другую функцию в основную
			}
		}(resp.Body)

		logger.Debugf("Resp:\n%s", resp)
		logger.Debugf("resp.Body:\n%s", resp.Body)

		respBody, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return err
		}

		logger.Debugf("Response:\n%s", respBody)

		resultGetLicenseIdByAnchor := new(ResultGetLicenseIdByAnchor)
		err = json.Unmarshal(respBody, resultGetLicenseIdByAnchor)
		if err != nil {
			return err
		}

		if resultGetLicenseIdByAnchor.Id == "" {
			return errors.New("не удалось получить ID лицензии XML")
		}

		x.Id = resultGetLicenseIdByAnchor.Id
		x.ExpirationDate = resultGetLicenseIdByAnchor.ExpirationDate

		if x.ExpirationDate != "" {
			x.ExpirationDateTimeFormat, err = time.Parse("2006-01-02T15:04:05", x.ExpirationDate)
			if err != nil {
				return errors.Wrapf(err, "failed time format %s", x.ExpirationDate)
			}
		}

		x.Qty = resultGetLicenseIdByAnchor.Qty
	}

	logger.Debugf("Id: %s", x.Id)
	logger.Debugf("ExpirationDate: %s", x.ExpirationDate)
	logger.Debugf("Qty: %f", x.Qty)

	return nil
}

func NewXmlInterface(UserName, Password, Token, ProductID, GUID string, RestCode int) (*xmlInterface, error) {
	logger := logging.GetLogger()
	logger.Println("Start NewXmlInterface")
	defer logger.Println("End NewXmlInterface")
	var err error

	if UserName == "" {
		return nil, errors.New("не указан UserName")
	}

	if Password == "" {
		return nil, errors.New("не указан Password")
	}

	if Token == "" {
		return nil, errors.New("не указан Token")
	}

	if ProductID == "" {
		return nil, errors.New("не указан ProductID")
	}

	if RestCode == 0 {
		return nil, errors.New("не указан RestCode")
	}

	xmlI = new(xmlInterface)
	xmlI.Usr, err = GenUsr(UserName, Password, Token)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка при генерации usr")
	}

	xmlI.Guid = GUID
	xmlI.ProductID = ProductID
	xmlI.RestCode = RestCode

	logger.Debugf("Anchor: %s", xmlI.Anchor)
	logger.Debugf("RestCode: %d", xmlI.RestCode)
	logger.Debugf("Guid: %s", xmlI.Guid)
	logger.Debugf("Usr: %s", xmlI.Usr)
	logger.Debugf("ProductID: %s", xmlI.ProductID)

	return xmlI, nil
}

func GetXmlInterface() (*xmlInterface, error) {
	logger := logging.GetLogger()
	logger.Println("Start GetXmlInterface")
	defer logger.Println("End GetXmlInterface")

	err := xmlI.GetLicenseIdByAnchor()
	if err != nil {
		return nil, errors.Wrap(err, "failed in xmlI.GetLicenseIdByAnchor()")
	}

	logger.Debugf("Id: %s", xmlI.Id)
	logger.Debugf("ExpirationDate: %s", xmlI.ExpirationDate)
	logger.Debugf("SeqNumber: %d", xmlI.SeqNumber)

	return xmlI, nil
}

func GenUsr(UserName, Password, Token string) (string, error) {

	logger := logging.GetLogger()
	logger.Println("GenUsr:>Start")
	defer logger.Println("GenUsr:>End")

	var err error

	hUserNameAndPassword := md5.New()
	logger.Debugf("Username: %s", UserName)
	logger.Debugf("Password: %s", Password)
	_, err = io.WriteString(hUserNameAndPassword, fmt.Sprintf("%s%s", UserName, Password))
	if err != nil {
		return "", err
	}
	userNameAndPassword := hUserNameAndPassword.Sum(nil)
	userNameAndPasswordString := fmt.Sprintf("%x", userNameAndPassword)
	logger.Debugf("UserName+Password MD5: %s", userNameAndPasswordString)

	hToken := md5.New()
	logger.Debugf("Token: %s", Token)
	_, err = io.WriteString(hToken, Token)
	if err != nil {
		return "", err
	}
	token := hToken.Sum(nil)
	tokenString := fmt.Sprintf("%x", token)
	logger.Debugf("Token MD5: %s", tokenString)

	var usr string
	usrNoneBase64 := fmt.Sprintf("%s;%s;%s", UserName, userNameAndPasswordString, tokenString)
	usr = base64.StdEncoding.EncodeToString([]byte(usrNoneBase64))

	return usr, nil
}
