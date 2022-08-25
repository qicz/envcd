/*
 * Licensed to the AcmeStack under one or more contributor license
 * agreements. See the NOTICE file distributed with this work for
 * additional information regarding copyright ownership.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package openapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/acmestack/envcd/internal/core/storage"
	"github.com/acmestack/envcd/internal/core/storage/dao"
	"github.com/acmestack/envcd/internal/pkg/constant"
	"github.com/acmestack/envcd/internal/pkg/entity"
	"github.com/acmestack/envcd/pkg/entity/result"
	"github.com/acmestack/godkits/array"
	"github.com/acmestack/godkits/gox/stringsx"
	"github.com/acmestack/pagehelper"
	"github.com/gin-gonic/gin"
)

type DictionaryDTO struct {
	UserId       int    `json:"userId" binding:"required"`
	ScopeSpaceId int    `json:"scopeSpaceId" binding:"required"`
	DictKey      string `json:"dictKey" binding:"required"`
	DictValue    string `json:"dictValue" binding:"required"`
	Version      string `json:"version" binding:"required"`
	State        string `json:"state" binding:"required"`
}

type dictionUpdateDTO struct {
	DictId    int    `json:"dictId" binding:"required"`
	DictValue string `json:"dictValue"`
	State     string `json:"state"`
}

func dictionary(storage *storage.Storage, dictionaryId *int, ginCtx *gin.Context) (*entity.Dictionary, error) {
	// get user id from gin context
	dictId := stringsx.ToInt(ginCtx.Param("dictionaryId"))
	if dictionaryId != nil {
		dictId = *dictionaryId
	}
	dict := entity.Dictionary{Id: dictId}
	dictionaries, err := dao.New(storage).SelectDictionary(dict, nil)
	if err != nil {
		return nil, err
	}
	if array.Empty(dictionaries) {
		return nil, nil
	}
	return &dictionaries[0], nil
}

// dictionary query single dictionary mapping
//  @receiver openapi common openapi
//  @param ginCtx gin context
func (openapi *Openapi) dictionary(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		dict, err := dictionary(openapi.storage, nil, ginCtx)
		if err != nil {
			return result.InternalFailure(err)
		}
		return result.Success(dict)
	})
}

// createDictionary create dictionary
//  @receiver openapi openapi
//  @param ginCtx gin context
func (openapi *Openapi) createDictionary(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		dictParams := &DictionaryDTO{}
		if err := ginCtx.ShouldBindJSON(dictParams); err != nil {
			fmt.Printf("Bind error, %v\n", err)
			return result.InternalFailure(err)
		}
		daoAction := dao.New(openapi.storage)
		// build dictionary with parameters
		dictionary := entity.Dictionary{
			UserId:       dictParams.UserId,
			ScopeSpaceId: dictParams.ScopeSpaceId,
			DictKey:      dictParams.DictKey,
			Version:      dictParams.Version,
			DictValue:    dictParams.DictValue,
			State:        dictParams.State,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		_, id, err := daoAction.InsertDictionary(dictionary)
		if err != nil {
			return result.InternalFailure(err)
		}
		path, PathErr := buildEtcdPath(daoAction, dictionary)
		if PathErr != nil {
			return result.Failure0(result.ErrorEtcdPath)
		}
		if stringsx.Empty(path) {
			return result.Failure0(result.NilExchangePath)
		}
		exchangeErr := openapi.exchange.Put(path, dictParams.DictValue)
		if exchangeErr != nil {
			return result.InternalFailure(exchangeErr)
		}
		openapi.doOperationLogging(dictParams.UserId, "create dictionary and insert into mysql and etcd")
		return result.Success(id)
	})
}

// updateDictionary update dictionary
//  @receiver openapi openapi
//  @param ginCtx gin context
func (openapi *Openapi) updateDictionary(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		dictParams := &dictionUpdateDTO{}
		if err := ginCtx.ShouldBindJSON(dictParams); err != nil {
			fmt.Printf("Bind error, %v\n", err)
			return result.InternalFailure(err)
		}
		daoAction := dao.New(openapi.storage)

		dictionary := entity.Dictionary{
			Id:        dictParams.DictId,
			DictValue: dictParams.DictValue,
			UpdatedAt: time.Now(),
		}
		// update dictionary
		_, updateDictErr := daoAction.UpdateDictionary(dictionary)
		if updateDictErr != nil {
			return result.InternalFailure(updateDictErr)
		}
		// update state
		ret := openapi.updateDictionaryState(dictParams.DictId, dictParams.State)
		if ret != nil {
			return ret
		}
		return result.Success(nil)
	})
}

// removeDictionary remove dictionary
//  @receiver openapi
//  @param ginCtx gin context
func (openapi *Openapi) removeDictionary(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		dict, err := dictionary(openapi.storage, nil, ginCtx)
		if err != nil {
			return result.InternalFailure(err)
		}
		if dict == nil {
			return result.Failure0(result.ErrorDictionaryNotExist)
		}
		daoAction := dao.New(openapi.storage)
		// set dictionaries state: deleted
		retId, delErr := daoAction.DeleteDictionary(*dict)
		if delErr != nil {
			return result.InternalFailure(delErr)
		}
		// delete etcd path
		path, etcdPathError := buildEtcdPath(daoAction, *dict)
		if etcdPathError != nil {
			return result.Failure0(result.ErrorEtcdPath)
		}
		if stringsx.Empty(path) {
			return result.Failure0(result.NilExchangePath)
		}
		if stringsx.NotEmpty(path) {
			exchangeErr := openapi.exchange.Remove(path)
			if exchangeErr != nil {
				return result.InternalFailure(exchangeErr)
			}
		}
		openapi.doOperationLogging(dict.UserId, "remove dictionaries from mysql and etcd")
		return result.Success(retId)
	})
}

func (openapi *Openapi) dictionaries(ginCtx *gin.Context) {
	openapi.response(ginCtx, nil, func() *result.EnvcdResult {
		pageNum := stringsx.ToInt(ginCtx.DefaultQuery("page", "1"))
		pageSize := stringsx.ToInt(ginCtx.DefaultQuery("pageSize", "20"))
		daoAction := dao.New(openapi.storage)
		ctx := pagehelper.C(context.Background()).PageWithCount(int64(pageNum-1), int64(pageSize), "").Build()
		dictionary, err := daoAction.SelectDictionary(entity.Dictionary{}, ctx)
		if err != nil {
			return result.InternalFailure(err)
		}
		pageInfo := pagehelper.GetPageInfo(ctx)
		return result.Success(PageListVO{
			Page:      pageInfo.Page + 1,
			PageSize:  pageInfo.PageSize,
			Total:     pageInfo.GetTotal(),
			TotalPage: pageInfo.GetTotalPage(),
			List:      dictionary,
		})
	})
}

// buildEtcdPath build etcd path
//  @param daoAction dao
//  @param dictionary
//  @return string path
//  @return error message
func buildEtcdPath(daoAction *dao.Dao, dictionary entity.Dictionary) (string, error) {
	// todo user name from jwt
	user, userErr := daoAction.SelectUser(entity.User{Id: dictionary.UserId})
	if userErr != nil {
		return "", userErr
	}
	scopeSpace, scopeSpaceErr := daoAction.SelectScopeSpace(entity.ScopeSpace{Id: dictionary.ScopeSpaceId})
	if scopeSpaceErr != nil {
		return "", scopeSpaceErr
	}
	// user and scopeSpace not exist
	if len(user) == 0 || len(scopeSpace) == 0 {
		return "", errors.New("user or spaceSpace not exist")
	}
	// build path
	build := stringsx.Builder{}
	// /scopeSpaceName/userName/dictKey, etc. /spring/moremind/userKey@version
	_, err := build.JoinString("/", scopeSpace[0].Name, "/", user[0].Name, "/", dictionary.DictKey)
	if err != nil {
		return "", err
	}
	return build.String(), nil
}

// updateDictionaryState update dictionary state
//  @receiver openapi openapi
//  @param dictId dict id
//  @param state updated state
//  @return *result.EnvcdResult
func (openapi *Openapi) updateDictionaryState(dictId int, state string) *result.EnvcdResult {
	daoAction := dao.New(openapi.storage)
	dictionaries, dictErr := daoAction.SelectDictionary(entity.Dictionary{Id: dictId}, nil)
	if dictErr != nil {
		return result.InternalFailure(dictErr)
	}
	if array.Empty(dictionaries) {
		return result.Failure0(result.ErrorDictionaryNotExist)
	}
	defaultDictionary := dictionaries[0]
	path, err := buildEtcdPath(daoAction, defaultDictionary)
	if stringsx.Empty(path) {
		return result.Failure0(result.NilExchangePath)
	}
	if err != nil {
		return result.Failure0(result.ErrorEtcdPath)
	}
	switch state {
	case constant.EnabledState:
		// case enabled, should generate path and put key and value
		if defaultDictionary.State != constant.EnabledState {
			exchangeErr := openapi.exchange.Put(path, defaultDictionary.DictValue)
			if exchangeErr != nil {
				return result.InternalFailure(exchangeErr)
			}
		}
		break
	case constant.DisabledState:
		// case disabled, should set state in mysql and delete dictionaries in etcd
	case constant.DeletedState:
		// case deleted, should set state in mysql and delete dictionaries in etcd
		if defaultDictionary.State == constant.DisabledState || defaultDictionary.State == constant.DeletedState {
			_, updateErr := daoAction.UpdateDictionary(entity.Dictionary{State: state})
			if updateErr != nil {
				return result.InternalFailure(updateErr)
			}
			exchangeErr := openapi.exchange.Remove(path)
			if exchangeErr != nil {
				return result.InternalFailure(exchangeErr)
			}
		}
		break
	default:
		return result.Failure0(result.ErrorNotExistState)
	}
	return nil
}
