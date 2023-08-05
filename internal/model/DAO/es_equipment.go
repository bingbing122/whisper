package dao

import (
	"errors"
	"fmt"
	"github.com/olivere/elastic/v7"
	"whisper/internal/model"
	"whisper/pkg/context"
	"whisper/pkg/es"
)

type ESEquipment interface {
	CreateIndex(ctx *context.Context) error
	Equipment2ES(ctx *context.Context, data []*model.ESEquipment) error
}

type ESEquipmentDAO struct {
	esClient *elastic.Client
}

func (dao *ESEquipmentDAO) CreateIndex(ctx *context.Context) error {

	// 索引是否存在
	var equip model.ESEquipment
	idxName := equip.GetIndexName()

	exists, err := es.ESClient.IndexExists(idxName).Do(ctx)
	if err != nil {
		return err
	}
	if !exists {
		// 创建索引
		var esEquip model.ESEquipment
		createIndex, err := es.ESClient.CreateIndex(idxName).Body(esEquip.GetMapping()).Do(ctx)
		if err != nil {
			return err
		}
		if !createIndex.Acknowledged {
			return errors.New(fmt.Sprintf("expected IndicesCreateResult.Acknowledged %v; got %v", true, createIndex.Acknowledged))
		}
	}
	return nil
}

func (dao *ESEquipmentDAO) Equipment2ES(ctx *context.Context, data []*model.ESEquipment) error {
	var equip model.ESEquipment
	idxName := equip.GetIndexName()
	// 导入数据
	for _, e := range data {
		_, err := es.ESClient.Index().Index(idxName).BodyJson(e).Id(e.ID).Do(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewESEquipmentDAO() *ESEquipmentDAO {
	return &ESEquipmentDAO{
		esClient: es.ESClient,
	}
}