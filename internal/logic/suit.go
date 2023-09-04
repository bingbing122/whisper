package logic

import (
	context2 "context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"whisper/internal/dto"
	"whisper/internal/logic/common"
	"whisper/internal/model"
	dao "whisper/internal/model/DAO"
	"whisper/internal/service"
	"whisper/pkg/context"
	"whisper/pkg/log"
	"whisper/pkg/redis"
)

var smu = &sync.Mutex{}

func BatchUpdateSuitEquip(ctx *context.Context) error {

	// 获取所有数据
	hd := dao.NewLOLHeroesDAO()
	version, err := hd.GetLOLHeroesMaxVersion()
	if err != nil {
		return err
	}
	heroes, err := hd.GetLOLHeroes(version.Version)
	if err != nil {
		return err
	}

	cancelCtx, cancelFunc := context2.WithCancel(ctx)
	defer cancelFunc()

	var (
		taskAll  = int32(len(heroes))
		taskSucc = int32(0)
		taskFail = int32(0)
		taskDone = int32(0)
		//mu       *sync.Mutex
		wg = &sync.WaitGroup{}
		ch = make(chan struct{}, 100)
	)

	for i, hero := range heroes {
		select {
		case <-cancelCtx.Done():
			break
		default:
			log.Logger.Info(ctx, ">>>>>>>>>>开始处理 hero:<<<<<<<<<<<", i, "/", hero.HeroId)
			ch <- struct{}{}
			wg.Add(1)

			go func(hero *model.LOLHeroes) {
				defer func() {
					<-ch
					wg.Done()
					atomic.AddInt32(&taskDone, 1)
				}()

				_, err2 := QuerySuitEquip(ctx, common.PlatformForLOL, hero.HeroId)
				// 任务执行失败，这个地方可以使用锁，也可以使用原子操作，优先原子操作
				if err2 != nil {
					atomic.AddInt32(&taskFail, 1)
					cancelFunc()
					log.Logger.Error(ctx, err2)
					return
				} else {
					atomic.AddInt32(&taskSucc, 1)
				}
			}(hero)

		}
	}

	wg.Wait()

	log.Logger.Info(ctx, fmt.Sprintf("处理了: %d 个任务", taskDone))
	log.Logger.Info(ctx, fmt.Sprintf("提前结束,执行出错: %d 个任务", taskFail))
	log.Logger.Info(ctx, fmt.Sprintf("成功执行了: %d 个任务", taskSucc))
	log.Logger.Info(ctx, fmt.Sprintf("剩余: %d 个任务待处理", taskAll-taskDone))

	return nil
}
func QuerySuitEquip(ctx *context.Context, platform int, heroId string) (*dto.ChampionFightData, error) {
	smu.Lock()
	defer smu.Unlock()

	fightData, err := getFightData(ctx, platform, heroId)
	if err != nil {
		return nil, errors.New("getFightData:" + err.Error())
	}

	// reload heroes_position 表
	err = updateHeroesPosition(ctx, platform, heroId, fightData)
	if err != nil {
		return nil, errors.New("updateHeroesPosition:" + err.Error())
	}

	// reload heroes_suit 表
	err = updateHeroesSuit(ctx, platform, heroId, fightData)
	if err != nil {
		return nil, errors.New("updateHeroesSuit:" + err.Error())
	}
	return fightData, nil
}

func getFightData(ctx *context.Context, platform int, heroId string) (*dto.ChampionFightData, error) {
	if platform == common.PlatformForLOL {
		fightData, err := service.ChampionFightData(ctx, heroId)
		if err != nil {
			return nil, err
		}
		for pos, posData := range fightData.List.ChampionLane {
			equipData := map[string]dto.Itemjson{}
			tmp := dto.ChampionLaneItem{}

			var err error
			err = json.Unmarshal([]byte(posData.Itemoutjson), &equipData)
			if err != nil {
				log.Logger.Warn(ctx, err, "heroid:", heroId)
			} else {
				tmp.Itemout = equipData
			}

			equipData = *new(map[string]dto.Itemjson)
			err = json.Unmarshal([]byte(posData.Core3itemjson), &equipData)
			if err != nil {
				log.Logger.Warn(ctx, err, "heroid:", heroId)
			} else {
				tmp.Core3item = equipData
			}

			equipData = *new(map[string]dto.Itemjson)
			err = json.Unmarshal([]byte(posData.Shoesjson), &equipData)
			if err != nil {
				log.Logger.Warn(ctx, err, "heroid:", heroId)
			} else {
				tmp.Shoes = equipData
			}

			var suits []dto.Itemjson
			err = json.Unmarshal([]byte(posData.Hold3), &suits)
			if err != nil {
				log.Logger.Warn(ctx, err, "heroid:", heroId)
			} else {
				tmp.Suits = suits
			}

			fightData.List.ChampionLane[pos] = tmp
		}

		return fightData, nil
	}

	return nil, nil
}
func updateHeroesPosition(ctx *context.Context, platform int, heroId string, fightData *dto.ChampionFightData) error {
	hpd := dao.NewHeroesPositionDAO()
	rows, err := hpd.Delete(map[string]interface{}{
		"heroId": heroId,
	})
	if err != nil {
		return errors.New("Delete HeroesPosition " + err.Error())
	}
	log.Logger.Info(ctx, "delete HeroesPosition rows:", rows, "heroId:", heroId)

	posData := make([]*model.HeroesPosition, 0, 3)
	for pos, _ := range fightData.List.ChampionFight {
		posData = append(posData, &model.HeroesPosition{
			HeroId:   heroId,
			Pos:      pos,
			Platform: platform,
			Version:  fightData.GameVer,
			FileTime: fightData.Date,
		})
	}
	if len(posData) == 0 {
		log.Logger.Warn(ctx, "posData is nil", "heroId:", heroId)
		return nil
	}
	rows, err = hpd.Add(posData)
	if err != nil {
		return errors.New("Add HeroesPosition " + err.Error() + ",heroId:" + heroId)
	}
	log.Logger.Info(ctx, "Add HeroesPosition rows:", rows, "heroId:", heroId)

	return nil
}
func updateHeroesSuit(ctx *context.Context, platform int, heroId string, fightData *dto.ChampionFightData) error {
	hpd := dao.NewHeroesSuitDAO()
	rows, err := hpd.Delete(map[string]interface{}{
		"heroId": heroId,
	})
	if err != nil {
		return errors.New("Delete HeroesSuit " + err.Error())
	}
	log.Logger.Info(ctx, "delete HeroesSuit rows:", rows, "heroId:", heroId)

	posData := make([]*model.HeroesSuit, 0)
	var m model.HeroesSuit
	for pos, pds := range fightData.List.ChampionLane {
		posCopy := pos
		for _, pdsd := range pds.Itemout {
			itemidArr := strings.Split(pdsd.Itemid, "&")
			itemids := strings.Join(itemidArr, ",")
			posData = append(posData, &model.HeroesSuit{
				HeroId:   heroId,
				Pos:      posCopy,
				Itemids:  itemids,
				Igamecnt: pdsd.Igamecnt,
				Wincnt:   pdsd.Wincnt,
				Winrate:  pdsd.Winrate,
				Allcnt:   pdsd.Allcnt,
				Showrate: pdsd.Showrate,
				Type:     m.TypeOut(),
				Platform: platform,
				Version:  fightData.GameVer,
				FileTime: fightData.Date,
			})
		}

		for _, pdsd := range pds.Core3item {
			itemidArr := strings.Split(pdsd.Itemid, "&")
			itemids := strings.Join(itemidArr, ",")
			posData = append(posData, &model.HeroesSuit{
				HeroId:   heroId,
				Pos:      posCopy,
				Itemids:  itemids,
				Igamecnt: pdsd.Igamecnt,
				Wincnt:   pdsd.Wincnt,
				Winrate:  pdsd.Winrate,
				Allcnt:   pdsd.Allcnt,
				Showrate: pdsd.Showrate,
				Type:     m.TypeCore(),
				Platform: platform,
				Version:  fightData.GameVer,
				FileTime: fightData.Date,
			})
		}

		for _, pdsd := range pds.Shoes {
			itemidArr := strings.Split(pdsd.Itemid, "&")
			itemids := strings.Join(itemidArr, ",")
			posData = append(posData, &model.HeroesSuit{
				HeroId:   heroId,
				Pos:      posCopy,
				Itemids:  itemids,
				Igamecnt: pdsd.Igamecnt,
				Wincnt:   pdsd.Wincnt,
				Winrate:  pdsd.Winrate,
				Allcnt:   pdsd.Allcnt,
				Showrate: pdsd.Showrate,
				Type:     m.TypeShoes(),
				Platform: platform,
				Version:  fightData.GameVer,
				FileTime: fightData.Date,
			})
		}

		for _, pdsd := range pds.Suits {
			itemidArr := strings.Split(pdsd.Itemid, "&")
			itemids := strings.Join(itemidArr, ",")
			posData = append(posData, &model.HeroesSuit{
				HeroId:   heroId,
				Pos:      posCopy,
				Itemids:  itemids,
				Igamecnt: pdsd.Igamecnt,
				Wincnt:   pdsd.Wincnt,
				Winrate:  pdsd.Winrate,
				Allcnt:   pdsd.Allcnt,
				Showrate: pdsd.Showrate,
				Type:     m.TypeOther(),
				Platform: platform,
				Version:  fightData.GameVer,
				FileTime: fightData.Date,
			})
		}
	}
	if len(posData) == 0 {
		log.Logger.Warn(ctx, "posData is nil", "heroId:", heroId)
		return nil
	}
	rows, err = hpd.Add(posData)
	if err != nil {
		return errors.New("Add HeroesSuit " + err.Error() + ",heroId:" + heroId)
	}
	log.Logger.Info(ctx, "Add HeroesSuit rows:", rows, "heroId:", heroId)

	return nil
}

func SuitData2Redis(ctx *context.Context) error {
	err := lolHeroes2Redis(ctx)
	if err != nil {
		return err
	}

	return nil
}
func lolHeroes2Redis(ctx *context.Context) error {
	hd := dao.NewLOLHeroesDAO()
	version, err := hd.GetLOLHeroesMaxVersion()
	if err != nil {
		return err
	}
	heroes, err := hd.GetLOLHeroes(version.Version)
	if err != nil {
		return err
	}

	// 获取全部装备
	ed := dao.NewLOLEquipmentDAO()
	eVersion, err := ed.GetLOLEquipmentMaxVersion()
	if err != nil {
		return err
	}
	equips, err := ed.GetLOLEquipment(eVersion.Version)
	if err != nil {
		return err
	}

	mequip := make(map[string]*model.LOLEquipment)
	for _, equip := range equips {
		key := fmt.Sprintf(redis.KeyCacheEquip, equip.Maps, equip.ItemId)
		value, _ := json.Marshal(equip)
		mequip[key] = equip
		redis.RDB.Set(ctx, key, value, time.Hour*2)
	}

	sd := dao.NewHeroesSuitDAO()

	cancelCtx, cancelFunc := context2.WithCancel(ctx)
	defer cancelFunc()

	var (
		taskAll  = int32(len(heroes))
		taskSucc = int32(0)
		taskFail = int32(0)
		taskDone = int32(0)
		//mu       *sync.Mutex
		wg = &sync.WaitGroup{}
		ch = make(chan struct{}, 10)
	)

	for i, hero := range heroes {
		select {
		case <-cancelCtx.Done():
			break
		default:
			log.Logger.Info(ctx, ">>>>>>>>>>开始处理 hero:<<<<<<<<<<<", i, "/", hero.HeroId)
			ch <- struct{}{}
			wg.Add(1)

			go func(hero *model.LOLHeroes) {
				defer func() {
					<-ch
					wg.Done()
					atomic.AddInt32(&taskDone, 1)
				}()

				equipForHero, err2 := sd.GetSuitForHero(hero.HeroId)
				if err2 != nil {
					atomic.AddInt32(&taskFail, 1)
					cancelFunc()
					log.Logger.Error(ctx, err2)
					return
				} else {

					hsm := make(map[string][]*model.HeroesSuit)
					for idx, equip := range equipForHero {
						hsm[equip.Pos] = append(hsm[equip.Pos], equipForHero[idx])
					}
					mhs := model.HeroesSuit{}

					eqs := make(map[string]dto.RecommendSuitEquip)
					for pos, posdata := range hsm {
						out := make([][]*dto.SuitData, 0)
						shoe := make([]*dto.SuitData, 0)
						core := make([][]*dto.SuitData, 0)
						other := make([]*dto.SuitData, 0)
						for _, data := range posdata {
							switch data.Type {
							case mhs.TypeShoes():
								key := fmt.Sprintf(redis.KeyCacheEquip, "召唤师峡谷", data.Itemids)
								if edata, ok := mequip[key]; !ok {
									continue
								} else {
									shoe = append(shoe, &dto.SuitData{
										ID:        cast.ToInt(data.Itemids),
										Name:      edata.Name,
										Icon:      edata.IconPath,
										Maps:      edata.Maps,
										Plaintext: edata.Plaintext,
										Desc:      edata.Description,
										Price:     cast.ToInt(edata.Total),
										Sell:      cast.ToInt(edata.Sell),
										Version:   edata.Version,

										Igamecnt: data.Igamecnt,
										Wincnt:   data.Wincnt,
										Winrate:  data.Winrate,
										Allcnt:   data.Allcnt,
										Showrate: data.Showrate,
									})
								}
							case mhs.TypeOther():
								key := fmt.Sprintf(redis.KeyCacheEquip, "召唤师峡谷", data.Itemids)
								if edata, ok := mequip[key]; !ok {
									continue
								} else {
									other = append(other, &dto.SuitData{
										ID:        cast.ToInt(data.Itemids),
										Name:      edata.Name,
										Icon:      edata.IconPath,
										Maps:      edata.Maps,
										Plaintext: edata.Plaintext,
										Desc:      edata.Description,
										Price:     cast.ToInt(edata.Total),
										Sell:      cast.ToInt(edata.Sell),
										Version:   edata.Version,

										Igamecnt: data.Igamecnt,
										Wincnt:   data.Wincnt,
										Winrate:  data.Winrate,
										Allcnt:   data.Allcnt,
										Showrate: data.Showrate,
									})
								}
							case mhs.TypeOut():
								ids := strings.Split(data.Itemids, ",")
								out2 := make([]*dto.SuitData, 0)
								for _, id := range ids {
									key := fmt.Sprintf(redis.KeyCacheEquip, "召唤师峡谷", id)
									if edata, ok := mequip[key]; !ok {
										continue
									} else {
										out2 = append(out2, &dto.SuitData{
											ID:        cast.ToInt(id),
											Name:      edata.Name,
											Icon:      edata.IconPath,
											Maps:      edata.Maps,
											Plaintext: edata.Plaintext,
											Desc:      edata.Description,
											Price:     cast.ToInt(edata.Total),
											Sell:      cast.ToInt(edata.Sell),
											Version:   edata.Version,

											Igamecnt: data.Igamecnt,
											Wincnt:   data.Wincnt,
											Winrate:  data.Winrate,
											Allcnt:   data.Allcnt,
											Showrate: data.Showrate,
										})
									}
								}
								out = append(out, out2)
							case mhs.TypeCore():
								ids := strings.Split(data.Itemids, ",")
								core2 := make([]*dto.SuitData, 0)
								for _, id := range ids {
									key := fmt.Sprintf(redis.KeyCacheEquip, "召唤师峡谷", id)
									if edata, ok := mequip[key]; !ok {
										continue
									} else {
										core2 = append(core2, &dto.SuitData{
											ID:        cast.ToInt(id),
											Name:      edata.Name,
											Icon:      edata.IconPath,
											Maps:      edata.Maps,
											Plaintext: edata.Plaintext,
											Desc:      edata.Description,
											Price:     cast.ToInt(edata.Total),
											Sell:      cast.ToInt(edata.Sell),
											Version:   edata.Version,

											Igamecnt: data.Igamecnt,
											Wincnt:   data.Wincnt,
											Winrate:  data.Winrate,
											Allcnt:   data.Allcnt,
											Showrate: data.Showrate,
										})
									}
								}
								core = append(core, core2)
							}
						}

						eqs[pos] = dto.RecommendSuitEquip{
							Out:   out,
							Shoe:  shoe,
							Core:  core,
							Other: other,
						}
					}

					jsonData, _ := json.Marshal(eqs)
					redis.RDB.HSet(ctx, redis.KeyCacheHeroEquip, hero.HeroId, jsonData)
					atomic.AddInt32(&taskSucc, 1)
				}
			}(hero)
		}
	}

	wg.Wait()

	log.Logger.Info(ctx, fmt.Sprintf("处理了: %d 个任务", taskDone))
	log.Logger.Info(ctx, fmt.Sprintf("提前结束,执行出错: %d 个任务", taskFail))
	log.Logger.Info(ctx, fmt.Sprintf("成功执行了: %d 个任务", taskSucc))
	log.Logger.Info(ctx, fmt.Sprintf("剩余: %d 个任务待处理", taskAll-taskDone))

	return nil
}

func GetHeroSuit(ctx *context.Context, heroID string) (map[string]dto.RecommendSuitEquip, error) {
	d := redis.RDB.HGet(ctx, redis.KeyCacheHeroEquip, heroID)

	var rs map[string]dto.RecommendSuitEquip
	err := json.Unmarshal([]byte(d.Val()), &rs)

	return rs, err
}