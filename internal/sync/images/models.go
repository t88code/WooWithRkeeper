package images

import "database/sql"

// imageJson Cтруктура Woo.Product.Images
type imageJson struct {
	Id      int    `json:"id"`
	Date    string `json:"date"`
	DateGmt string `json:"date_gmt"`
	Guid    struct {
		Rendered string `json:"rendered"`
		Raw      string `json:"raw"`
	} `json:"guid"`
	Modified    string `json:"modified"`
	ModifiedGmt string `json:"modified_gmt"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Link        string `json:"link"`
	Title       struct {
		Raw      string `json:"raw"`
		Rendered string `json:"rendered"`
	} `json:"title"`
	Author            int           `json:"author"`
	CommentStatus     string        `json:"comment_status"`
	PingStatus        string        `json:"ping_status"`
	Template          string        `json:"template"`
	Meta              []interface{} `json:"meta"`
	PermalinkTemplate string        `json:"permalink_template"`
	GeneratedSlug     string        `json:"generated_slug"`
	Acf               []interface{} `json:"acf"`
	Description       struct {
		Raw      string `json:"raw"`
		Rendered string `json:"rendered"`
	} `json:"description"`
	Caption struct {
		Raw      string `json:"raw"`
		Rendered string `json:"rendered"`
	} `json:"caption"`
	AltText      string `json:"alt_text"`
	MediaType    string `json:"media_type"`
	MimeType     string `json:"mime_type"`
	MediaDetails struct {
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		File     string `json:"file"`
		Filesize int    `json:"filesize"`
		Sizes    struct {
			Medium struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"medium"`
			Thumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"thumbnail"`
			WoocommerceThumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				Uncropped bool   `json:"uncropped"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"woocommerce_thumbnail"`
			WoocommerceSingle struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"woocommerce_single"`
			WoocommerceGalleryThumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"woocommerce_gallery_thumbnail"`
			QuickViewImageSize struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"quick_view_image_size"`
			Full struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"full"`
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
	Post              interface{} `json:"post"`
	SourceUrl         string      `json:"source_url"`
	MissingImageSizes []string    `json:"missing_image_sizes"`
	Links             struct {
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
		WpActionUnfilteredHtml []struct {
			Href string `json:"href"`
		} `json:"wp:action-unfiltered-html"`
		WpActionAssignAuthor []struct {
			Href string `json:"href"`
		} `json:"wp:action-assign-author"`
		Curies []struct {
			Name      string `json:"name"`
			Href      string `json:"href"`
			Templated bool   `json:"templated"`
		} `json:"curies"`
	} `json:"_links"`
}

// ImageSelectJoin Структура для SELECT Db.Image LEFT JOIN Db.ImageFile
type ImageSelectJoin struct {
	IdentRK  int            `db:"IdentRK"`
	Name     sql.NullString `db:"Name"`     //todo перепроверить переменные
	Pos      sql.NullInt32  `db:"Pos"`      //todo проверить это 0 или null, когда не указан
	Status   sql.NullString `db:"Status"`   //todo Status https://docs.google.com/spreadsheets/d/1oZ7jDxDHMfHvsLfN90HYmV3Cdj6uO_hS4MVBGJzuKqU/edit#gid=1756609729
	ModTime  sql.NullString `db:"ModTime"`  //todo перепроверить переменные
	IdentWOO sql.NullInt32  `db:"IdentWOO"` //todo проверить это 0 или null, когда не указан
}
