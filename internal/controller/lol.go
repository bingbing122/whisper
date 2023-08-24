package controller

import (
	"strings"
	"whisper/internal/logic"
	"whisper/pkg/errors"

	"whisper/pkg/context"
)

type ReqEquipment struct {
	Platform int `form:"platform" json:"platform" binding:"-"`
	//Platform int `form:"platform" json:"platform" binding:"required"`
}

func Equipment(ctx *context.Context) {

	req := &ReqEquipment{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	equip, err := logic.QueryEquipments(ctx, req.Platform)

	ctx.Reply(equip, errors.New(err))
}

type ReqHeroes struct {
	Platform int `form:"platform" json:"platform" binding:"-"`
}

func Heroes(ctx *context.Context) {

	req := &ReqHeroes{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	equip, err := logic.QueryHeroes(ctx, req.Platform)

	ctx.Reply(equip, errors.New(err))
}

type ReqHeroesAttribute struct {
	Platform int    `form:"platform" json:"platform" binding:"-"`
	HeroID   string `form:"hero_id" json:"hero_id" binding:"required"`
}

func HeroesAttribute(ctx *context.Context) {

	req := &ReqHeroesAttribute{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	attr, err := logic.HeroAttribute(ctx, req.HeroID, req.Platform)

	ctx.Reply(attr, errors.New(err))
}

type ReqRune struct {
	Platform int `form:"platform" json:"platform" binding:"-"`
}

func Rune(ctx *context.Context) {

	req := &ReqRune{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	equip, err := logic.QueryRune(ctx, req.Platform)

	ctx.Reply(equip, errors.New(err))
}

type ReqRuneType struct {
	Platform int `form:"platform" json:"platform" binding:"-"`
}

func RuneType(ctx *context.Context) {

	req := &ReqRuneType{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	equip, err := logic.QueryRuneType(ctx, req.Platform)

	ctx.Reply(equip, errors.New(err))
}

type ReqSkill struct {
	Platform int `form:"platform" json:"platform" binding:"-"`
}

func Skill(ctx *context.Context) {

	req := &ReqSkill{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	equip, err := logic.QuerySkill(ctx, req.Platform)

	ctx.Reply(equip, errors.New(err))
}

type ReqEquipExtract struct {
	Platform int `form:"platform" json:"platform" binding:"-"`
}

func EquipExtract(ctx *context.Context) {

	req := &ReqEquipExtract{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	words := logic.ExtractKeyWords(ctx, req.Platform)

	ctx.Reply(words, errors.New(nil))
}

type ReqEquipFilter struct {
	Platform int      `form:"platform" json:"platform" binding:"-"`
	Keywords []string `json:"keywords"`
}

func EquipFilter(ctx *context.Context) {

	req := &ReqEquipFilter{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	keywords := make([]string, 0)
	for _, filter := range req.Keywords {
		keywords = append(keywords, strings.Split(filter, ",")...)
	}

	equips, err := logic.FilterKeyWords(ctx, keywords, req.Platform)

	ctx.Reply(equips, errors.New(err))
}
