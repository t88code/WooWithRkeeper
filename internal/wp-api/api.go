package wp_api

import (
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

func GetMediaAll() (result []MediaJson, err error) {
	logger := logging.GetLogger()
	logger.Debug("Start GetMediaAll")
	defer logger.Debug("End GetMediaAll")

	cfg := config.GetConfig()
	url := fmt.Sprintf("%s/wp-json/wp/v2/media", cfg.WOOCOMMERCE.URL)

	client := &http.Client{}

	page := 0
ForBreak:
	for {
		page++
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "http.NewRequest")
		}
		req.SetBasicAuth(cfg.WOOCOMMERCE.User, cfg.WOOCOMMERCE.Password)
		q := req.URL.Query()
		q.Add("after", cfg.IMAGESYNC.DateAfter)
		q.Add("per_page", "100")
		q.Add("page", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()

		resp, err := client.Do(req)
		if err != nil {
			return nil, errors.Wrapf(err, "failed in client.Do")
		}
		switch resp.Status {
		case "200 OK":
			content, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, errors.Wrapf(err, "ioutil.ReadAll")
			}
			var i []MediaJson
			err = json.Unmarshal(content, &i)
			if err != nil {
				return nil, errors.Wrap(err, "Не удалось выполнить Unmarshal")
			} else {
				if len(i) == 0 {
					break ForBreak
				}
				result = append(result, i...)
			}
		case "400 Bad Request":
			content, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, errors.Wrapf(err, "ioutil.ReadAll")
			}
			var message mediaError
			err = json.Unmarshal(content, &message)
			if err != nil {
				return nil, errors.Wrap(err, "Не удалось выполнить Unmarshal")
			} else {
				if message.Code == "rest_post_invalid_page_number" {
					break ForBreak
				} else {
					return nil, errors.New(fmt.Sprintf("400 Not Found; error: %s", message.Message))
				}
			}
		default:
			return nil, errors.New(fmt.Sprintf("Ошибка при попытке получить media. Status: %s", resp.Status))
		}
	}

	return result, nil
}

type mediaError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Status int `json:"status"`
	} `json:"data"`
}

type MediaJson struct {
	Id      int    `json:"id"`
	Date    string `json:"date"`
	DateGmt string `json:"date_gmt"`
	Guid    struct {
		Rendered string `json:"rendered"`
	} `json:"guid"`
	Modified    string `json:"modified"`
	ModifiedGmt string `json:"modified_gmt"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Link        string `json:"link"`
	Title       struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Author        int           `json:"author"`
	CommentStatus string        `json:"comment_status"`
	PingStatus    string        `json:"ping_status"`
	Template      string        `json:"template"`
	Meta          []interface{} `json:"meta"`
	Acf           []interface{} `json:"acf"`
	Description   struct {
		Rendered string `json:"rendered"`
	} `json:"description"`
	Caption struct {
		Rendered string `json:"rendered"`
	} `json:"caption"`
	AltText      string `json:"alt_text"`
	MediaType    string `json:"media_type"`
	MimeType     string `json:"mime_type"`
	MediaDetails struct {
		Width  int    `json:"width"`
		Height int    `json:"height"`
		File   string `json:"file"`
		Sizes  struct {
			Medium struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"medium,omitempty"`
			Thumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"thumbnail,omitempty"`
			WoocommerceThumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				Uncropped bool   `json:"uncropped"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"woocommerce_thumbnail,omitempty"`
			WoocommerceGalleryThumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"woocommerce_gallery_thumbnail,omitempty"`
			ShopCatalog struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"shop_catalog,omitempty"`
			ShopThumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"shop_thumbnail,omitempty"`
			Full struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"full,omitempty"`
		} `json:"sizes"`
		ImageMeta struct {
			Aperture         string        `json:"aperture"`
			Credit           string        `json:"credit"`
			Camera           string        `json:"camera"`
			Caption          string        `json:"caption"`
			CreatedTimestamp string        `json:"created_timestamp"`
			Copyright        string        `json:"copyright"`
			FocalLength      string        `json:"focal_length"`
			Iso              string        `json:"iso"`
			ShutterSpeed     string        `json:"shutter_speed"`
			Title            string        `json:"title"`
			Orientation      string        `json:"orientation"`
			Keywords         []interface{} `json:"keywords"`
		} `json:"image_meta"`
	} `json:"media_details"`
	Post      *int   `json:"post"`
	SourceUrl string `json:"source_url"`
	Links     struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
		Collection []struct {
			Href string `json:"href"`
		} `json:"collection"`
		About []struct {
			Href string `json:"href"`
		} `json:"about"`
		Author []struct {
			Embeddable bool   `json:"embeddable"`
			Href       string `json:"href"`
		} `json:"author"`
		Replies []struct {
			Embeddable bool   `json:"embeddable"`
			Href       string `json:"href"`
		} `json:"replies"`
	} `json:"_links"`
}
