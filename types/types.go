package types

type Player struct {
	AccountID   int64  `json:"account_id"`
    Nickname    string `json:"nickname"` //атрибуты для идентификации

	DotaMMR     int    `json:"dota_mmr"`
    Role        string `json:"role"`

	TournamentPTS float64 `json:"t_pts"`
    MatchesPlayed int     `json:"matches"`
    AverageFa     float64 `json:"avg_fa"` //6 атрибутов для аналитики

	GoldTimes []int `json:"gold_t"`
	Creeps    int   `json:"last_hits"`
	Kills     int   `json:"kills"`
	Deaths    int   `json:"deaths"`
	Assists   int   `json:"assists"`
	Damage    int   `json:"hero_damage"`
	XPM       int   `json:"xp_per_min"`
	HeroID 	  int   `json:"hero_id"`
	HeroName string 
}

type Match struct {
	Duration int `json:"duration"`
	TowersR int `json:"tower_status_radiant"`
	TowersD int `json:"tower_status_dire"`
	Players []Player `json:"players"`
	GoldAdv []int `json:"radiant_gold_adv"`
	IsRadWin bool `json:"radiant_win"`
}