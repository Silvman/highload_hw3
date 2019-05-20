package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"highload_hw3/models"
	"highload_hw3/pkg/responses"
	"highload_hw3/storage"
	"net/http"
	"strings"
)

func (srv *Server) saveScore(w http.ResponseWriter, r *http.Request) {
	var scoreReq models.Score
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		srv.log.Warnln("can't read request from body:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(body, &scoreReq); err != nil {
		srv.log.Warnln("can't unmarshal request:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(getUserID(r))
	fmt.Println(scoreReq.Score)
	if err := srv.users.UpdateScore(int(getUserID(r)), scoreReq.Score); err != nil {
		srv.log.Warnln("can't update user score into db:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	responses.Write(w, http.StatusOK, scoreReq)
}

func (srv *Server) getShopPositions(w http.ResponseWriter, r *http.Request) {
	upgrades, err := srv.game.GetAllUpgrades()
	if err != nil {
		srv.log.Println("can't get upgrades from db:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	responses.Write(w, http.StatusOK, upgrades)
}

func (srv *Server) buyUpgrades(w http.ResponseWriter, r *http.Request) {
	var upgradeReq models.Upgrade
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		srv.log.Warnln("can't read request from body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(body, &upgradeReq); err != nil {
		srv.log.Warnln("can't unmarshal request", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := srv.game.BuyUpgrade(int(getUserID(r)), upgradeReq.Id); err != nil {
		if strings.EqualFold(storage.ErrNeedMoreGold, err.Error()) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			srv.log.Warnln("can't buy upgrade:", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (srv *Server) getAchievement(w http.ResponseWriter, r *http.Request) {
	// TODO add achievement
	responses.Write(w, http.StatusMethodNotAllowed, nil)
}
