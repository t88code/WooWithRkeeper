package images

import (
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/database"
	wp_api "WooWithRkeeper/internal/wp-api"
	"WooWithRkeeper/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
)

type imageFile struct {
	Name     string
	ModTime  string
	IdentWOO int
}

func GetImageFilesInDbMap(db *sqlx.DB) (map[string]imageFile, error) {

	logger := logging.GetLogger()
	logger.Debug("Start GetImageFilesInDbMap")
	defer logger.Debug("End GetImageFilesInDbMap")
	var err error

	// Формирование imageFilesInDbMap
	// -получаем записи из DB
	// -создаем map[string]imageFile с проверкой на дубли
	var imageFilesInDb []database.ImageFile
	query := "SELECT * FROM ImageFile"
	err = db.Select(&imageFilesInDb, query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed SELECT to dbsqlite; query %s", query)
	}

	// создаем map[string]imageFile с проверкой на дубли
	imageFilesInDbMap := make(map[string]imageFile, 0)
	for _, imageFileInDb := range imageFilesInDb {
		if imageFileInDb.Name.Valid && imageFileInDb.ModTime.Valid && imageFileInDb.IdentWOO.Valid {
			if _, found := imageFilesInDbMap[imageFileInDb.Name.String]; !found {
				imageFilesInDbMap[imageFileInDb.Name.String] = imageFile{
					Name:     imageFileInDb.Name.String,
					ModTime:  imageFileInDb.ModTime.String,
					IdentWOO: int(imageFileInDb.IdentWOO.Int32),
				}
			} else {
				logger.Error("Запись в DB дублируется. Некорректное поведение")
				return nil, errors.New(fmt.Sprintf("Запись в DB дублируется. Некорректное поведение: %v", imageFileInDb))
			}
		}
	}

	logger.Debugf("imageFilesInDbMap сформирован: len=%d", len(imageFilesInDbMap))
	return imageFilesInDbMap, nil

}

func GetImageFilesInDirMap() (map[string]fs.FileInfo, error) {

	logger := logging.GetLogger()
	logger.Debug("Start GetImageFilesInDirMap")
	defer logger.Debug("End GetImageFilesInDirMap")
	var err error
	cfg := config.GetConfig()

	// Формирование imageFilesInDirMap
	// -получаем файлы из папки
	// -создаем imageFilesInDirMap
	files, err := ioutil.ReadDir(cfg.IMAGESYNC.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed in ioutil.ReadDir")
	}

	// создаем map[string]File
	imageFilesInDirMap := make(map[string]fs.FileInfo, 0)
	for _, file := range files {
		match, _ := regexp.MatchString(".jpg$", file.Name())
		if match {
			if imageFile, found := imageFilesInDirMap[file.Name()]; !found {
				imageFilesInDirMap[file.Name()] = file
			} else {
				logger.Error("Файл в папке дублируется. Некорректное поведение")
				return nil, errors.New(fmt.Sprintf("Файл в папке дублируется. Некорректное поведение: %v", imageFile))
			}
		}
	}

	logger.Debugf("imageFilesInDirMap сформирован: len=%d", len(imageFilesInDirMap))
	return imageFilesInDirMap, nil
}

func GetImageFilesInWooMap() (map[int]wp_api.MediaJson, error) {
	logger := logging.GetLogger()
	logger.Debug("Start GetImageFilesInWooMap")
	defer logger.Debug("End GetImageFilesInWooMap")
	var err error

	// Формирование imageFilesInWooMap
	// -получаем список файлов из WOO
	// -создаем imageFilesInWooMap
	mediaJsons, err := wp_api.GetMediaAll()
	if err != nil {
		return nil, errors.Wrap(err, "failed in wp_api.GetMediaAll")
	}

	// создаем map[int]wp_api.MediaJson
	imageFilesInWooMap := make(map[int]wp_api.MediaJson)
	for _, mediaJson := range mediaJsons {
		imageFilesInWooMap[mediaJson.Id] = mediaJson
	}

	logger.Debugf("imageFilesInWooMap сформирован: len=%d", len(imageFilesInWooMap))
	return imageFilesInWooMap, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandlerImageFileToDb 1 этап - Синхронизация файлов картинок из папки с Woo.Media/DB.ImageFiles
// Картинки закачиваются в Woo.Media и присваивается Woo идентификатор в DB.ImageFiles
func HandlerImageFileToDb(db *sqlx.DB) error {

	logger := logging.GetLogger()
	logger.Debug("Start HandlerImageFileToDb")
	defer logger.Debug("End HandlerImageFileToDb")
	var err error

	// Общий алгоритм:
	// -формирование imageFilesInDbMap = получили то что в DB
	// -формирование imageFilesInDirMap = получили то что в DIR
	// -формирование imageFilesInWooMap = получили то что в WOO
	// -удаляем лишние записи из DB и WOO
	// -сверяем даты

	imageFilesInDbMap, err := GetImageFilesInDbMap(db)
	if err != nil {
		return errors.Wrap(err, "failed in GetImageFilesInDbMap")
	}

	imageFilesInDirMap, err := GetImageFilesInDirMap()
	if err != nil {
		return errors.Wrap(err, "failed in GetImageFilesInDirMap")
	}

	imageFilesInWooMap, err := GetImageFilesInWooMap()
	if err != nil {
		return errors.Wrap(err, "failed in imageFilesInWooMap")
	}

	// удаляем лишние записи из DB и WOO
	for _, imageFileInDb := range imageFilesInDbMap {
		logger.Debug("---------------------------------------------")
		logger.Debugf("Найдена запись в DB: Name=%s, ModTime=%s, IdentWOO=%d",
			imageFileInDb.Name, imageFileInDb.ModTime, imageFileInDb.IdentWOO)
		if _, found := imageFilesInDirMap[imageFileInDb.Name]; !found {
			logger.Debugf("Файл не найден, но имеется в DB. Удаляем файл из WOO и запись из DB: Name=%s", imageFileInDb.Name)

			if _, found := imageFilesInWooMap[imageFileInDb.IdentWOO]; found {
				err = DeleteImageInWoo(imageFileInDb.IdentWOO)
				if err != nil {
					if err.Error() == "Картинка не удалена. Status: 404 Not Found" {
						logger.Debug("Картинка не удалена, не найдена в WOO.Media")
					} else {
						return errors.Wrap(err, "failed in DeleteImageInWoo")
					}
				}
				delete(imageFilesInWooMap, imageFileInDb.IdentWOO)
			} else {
				logger.Debug("Удаление не требуется. Картинка не найдена в WOO")
			}

			err := DeleteImageInDB(db, imageFileInDb.Name)
			if err != nil {
				return err
			}
		} else {
			logger.Debug("Файл картинки найден в папке. Удалять запись в DB не требуется")
		}
	}

	// Алгоритм сверки:
	// -если картинка в DB есть
	// --если дата картинки не совпадает, то удаляем и закачиваем в WOO и обновляем в DB
	// --если дата картинки совпадает, то проверяем наличие в WOO
	// ---если картинка не найдена в WOO то закачиваем и обновляем в DB
	// -если картинки в DB нет, то закачиваем в WOO и DB
	for _, imageFileInDir := range imageFilesInDirMap {
		logger.Debug("---------------------------------------------")
		logger.Debugf("Файл в папке: Name=%s, ModTime=%s", imageFileInDir.Name(), imageFileInDir.ModTime())
		if imageFileInDb, found := imageFilesInDbMap[imageFileInDir.Name()]; found {
			logger.Debugf("Найдена запись в DB: Name=%s, ModTime=%s, IdentWOO=%d",
				imageFileInDb.Name, imageFileInDb.ModTime, imageFileInDb.IdentWOO)
			logger.Debug("Сравнить и если требуется, то обновить")
			if imageFileInDir.ModTime().Format(time.RFC3339) != imageFileInDb.ModTime {
				logger.Debug("Обновление требуется. Дата различается. Удаляем и закачиваем снова")
				err = DeleteImageInWoo(imageFileInDb.IdentWOO)
				if err != nil {
					if err.Error() == "Картинка не удалена. Status: 404 Not Found" {
						logger.Debug("Картинка не удалена, не найдена в WOO.Media")
					} else {
						return errors.Wrap(err, "failed in DeleteImageInWoo")
					}
				}
				identWOO, err := UploadImageToWoo(imageFileInDir.Name())
				if err != nil {
					return errors.Wrapf(err, "failed in UploadImageToWoo(%s)", imageFileInDir.Name())
				} else {
					logger.Debug("Картинка успешно закачена в WOO")
					err := UpdateImageInDb(db, imageFileInDir.Name(), imageFileInDir.ModTime().Format(time.RFC3339), identWOO)
					if err != nil {
						return errors.Wrap(err, "failed in UpdateImageInDb")
					} else {
						logger.Debug("Картинка успешно обновлена в DB")
					}
				}
			} else {
				logger.Debugf("Дата совпадает. Проверяем наличие картинки в WOO.Media. IdentWOO=%d", imageFileInDb.IdentWOO)
				if _, found := imageFilesInWooMap[imageFileInDb.IdentWOO]; found {
					logger.Debug("Картинка найдена в WOO. Закачка не требуется.")
				} else {
					logger.Debugf("Картинка(id=%d) не найдена в WOO.Media. Закачиваем.", imageFileInDb.IdentWOO)
					identWOO, err := UploadImageToWoo(imageFileInDir.Name())
					if err != nil {
						return errors.Wrapf(err, "failed in UploadImageToWoo(%s)", imageFileInDir.Name())
					} else {
						logger.Debug("Картинка успешно закачена в WOO")
						err := UpdateImageInDb(db, imageFileInDir.Name(), imageFileInDir.ModTime().Format(time.RFC3339), identWOO)
						if err != nil {
							return errors.Wrap(err, "failed in UpdateImageInDb")
						} else {
							logger.Debug("Картинка успешно обновлена в DB")
						}
					}
				}
			}
		} else {
			logger.Debug("Не найдена запись в DB. Закачиваем картинку в WOO и добавляем запись в DB")
			identWOO, err := UploadImageToWoo(imageFileInDir.Name())
			if err != nil {
				return errors.Wrapf(err, "failed in UploadImageToWoo(%s)", imageFileInDir.Name())
			} else {
				logger.Debug("Картинка успешно закачена в WOO")
				err := AddImageToDb(db, imageFileInDir.Name(), imageFileInDir.ModTime().Format(time.RFC3339), identWOO)
				if err != nil {
					return errors.Wrap(err, "failed in AddImageToDb")
				} else {
					logger.Debug("Картинка успешно добавлена в DB")
				}
			}
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// DeleteImageInWoo Удалить картинку из WOO.Media
// Используется в 1 этапе синхронизации картинок
func DeleteImageInWoo(id int) error {
	logger := logging.GetLogger()
	logger.Debug("Start DeleteImageInWoo")
	defer logger.Debug("End DeleteImageInWoo")

	cfg := config.GetConfig()
	url := fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, id)

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return errors.Wrapf(err, "http.NewRequest")
	}

	q := req.URL.Query()
	q.Add("force", "true")
	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth(cfg.WOOCOMMERCE.User, cfg.WOOCOMMERCE.Password)
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed in client.Do")
	}
	switch resp.Status {
	case "200 OK":
		logger.Debug("Картинка удалена успешно")
		return nil
	default:
		return errors.New(fmt.Sprintf("Картинка не удалена. Status: %s", resp.Status))
	}
}

// DeleteImageInDB Удалить картинку из DB.ImageFile
// Используется в 1 этапе синхронизации картинок
func DeleteImageInDB(db *sqlx.DB, name string) error {
	logger := logging.GetLogger()
	logger.Debug("Start DeleteImageInDB")
	defer logger.Debug("End DeleteImageInDB")

	var err error
	var query string

	query = "DELETE FROM ImageFile WHERE Name=?"

	result := db.MustExec(query, name)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed in result.RowsAffected(); query=\n%s", query)
	} else {
		logger.Debugf("Успешно удалено строк: %d", rowsAffected)
		if rowsAffected > 1 {
			logger.Warningf("Удалено строк более чем 1: %d", rowsAffected)
		}
	}

	return nil
}

// UploadImageToWoo Закачать файл картинки из папки в Woo.Media
// Используется в 1 этапе синхронизации картинок
func UploadImageToWoo(filename string) (int, error) {
	logger := logging.GetLogger()
	logger.Debug("Start UploadImageToWoo")
	defer logger.Debug("End UploadImageToWoo")

	cfg := config.GetConfig()
	url := fmt.Sprintf("%s/wp-json/wp/v2/media", cfg.WOOCOMMERCE.URL)
	path := fmt.Sprintf("%s/%s", cfg.IMAGESYNC.Path, filename)

	data, err := os.Open(path)
	if err != nil {
		return 0, errors.Wrapf(err, "failed in os.Open")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return 0, errors.Wrapf(err, "http.NewRequest")
	}
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("Content-Disposition", fmt.Sprintf(`form-data; filename="%s"`, filename))

	req.SetBasicAuth(cfg.WOOCOMMERCE.User, cfg.WOOCOMMERCE.Password)
	resp, err := client.Do(req)
	if err != nil {
		return 0, errors.Wrapf(err, "failed in client.Do")
	}
	switch resp.Status {
	case "201 Created":
		logger.Debug("Картинка закачана успешно")
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, errors.Wrapf(err, "ioutil.ReadAll")
		}
		i := new(imageJson)
		err = json.Unmarshal(content, i)
		if err != nil {
			return 0, errors.Wrap(err, "Не удалось выполнить Unmarshal")
		} else {
			return i.Id, nil
		}
	default:
		return 0, errors.New(fmt.Sprintf("Картинка не закачана. Status: %s", resp.Status))
	}
}

// AddImageToDb Добавить запись в DB
// Используется в 1 этапе синхронизации картинок
func AddImageToDb(db *sqlx.DB, name, modtime string, identwoo int) error {
	logger := logging.GetLogger()
	logger.Debug("Start AddImageToDb")
	defer logger.Debug("End AddImageToDb")

	var err error

	query := "INSERT INTO ImageFile (Name, ModTime, IdentWOO) VALUES ($1, $2, $3);"
	logger.Debugf("INSERT:\n%s(%s, %s, %d)", query, name, modtime, identwoo)
	result := db.MustExec(query, name, modtime, identwoo)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed INSERT to dbsqlite; query:\n%s(%s, %s, %d)", query, name, modtime, identwoo)
	}
	if rowsAffected == 0 {
		return errors.New("rowsAffected=0")
	} else {
		logger.Debugf("Успешно записано строк: %d", rowsAffected)
		return nil
	}
}

// UpdateImageInDb Обновить запись в DB поля WOO_ID/ModTime
// Используется в 1 этапе синхронизации картинок
func UpdateImageInDb(db *sqlx.DB, name, modtime string, identwoo int) error {
	logger := logging.GetLogger()
	logger.Debug("Start UpdateImageInDb")
	defer logger.Debug("End UpdateImageInDb")

	var err error

	query := "UPDATE ImageFile SET IdentWOO=:IdentWOO, ModTime=:ModTime WHERE Name=:Name;"
	logger.Debugf("UPDATE:\n%s(%d, %s, %s)", query, identwoo, name, modtime)
	_, err = db.NamedExec(query,
		map[string]interface{}{
			"Name":     name,
			"IdentWOO": identwoo,
			"ModTime":  modtime,
		})
	if err != nil {
		return errors.Wrap(err, "failed in db.NamedExec")
	} else {
		return nil
	}
}
