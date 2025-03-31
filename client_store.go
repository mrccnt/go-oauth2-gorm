package oauth2gorm

import (
	"context"
	"encoding/json"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"gorm.io/gorm"
	"io"
	"os"
)

type ClientStoreItem struct {
	ID     string
	Secret string `gorm:"type:varchar(512)"`
	Domain string `gorm:"type:varchar(512)"`
	Data   string `gorm:"type:text"`
}

func NewClientStore(table string, db *gorm.DB) *ClientStore {
	store := &ClientStore{
		db:        db,
		tableName: table,
		stdout:    os.Stderr,
	}
	if !db.Migrator().HasTable(store.tableName) {
		if err := db.Table(store.tableName).Migrator().CreateTable(&ClientStoreItem{}); err != nil {
			panic(err)
		}
	}
	return store
}

type ClientStore struct {
	tableName string
	db        *gorm.DB
	stdout    io.Writer
}

func (s *ClientStore) toClientInfo(data []byte) (oauth2.ClientInfo, error) {
	var cm models.Client
	err := json.Unmarshal(data, &cm)
	return &cm, err
}

func (s *ClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}

	var item ClientStoreItem
	err := s.db.WithContext(ctx).Table(s.tableName).Limit(1).Find(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return s.toClientInfo([]byte(item.Data))
}

func (s *ClientStore) Create(ctx context.Context, info oauth2.ClientInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	item := &ClientStoreItem{
		ID:     info.GetID(),
		Secret: info.GetSecret(),
		Domain: info.GetDomain(),
		Data:   string(data),
	}

	return s.db.WithContext(ctx).Table(s.tableName).Create(item).Error
}
