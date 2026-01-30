package facoef

import (
	"math/bits"
	"math"
	"coefbot/types"
)

func SumPlayers(ps []types.Player, dur int) (float64, []float64) {
	var sum float64
	var players []float64
	sum = 0
	for _, p := range ps {
		logC := math.Log(float64(p.GoldTimes[len(p.GoldTimes)-1]) + 1)
		creepC := math.Cbrt(float64(p.Creeps + 1))
		KDAC := math.Pow(1 + (float64(p.Kills) + types.AssistsConst * float64(p.Assists)) / (float64(p.Deaths) + 1), types.KDAExp)
		XPMC := 1 - math.Pow(math.E, -(float64(p.XPM)/types.XPMConst))
		DMGC := math.Tanh(float64(p.Damage)/types.DmgConst/float64(dur))
		sum += logC*creepC*KDAC*XPMC*DMGC
		players = append(players, logC*creepC*KDAC*XPMC*DMGC)
	}
	return sum, players
}

func networthStandardDeviation(p []types.Player) float64 {
   var sum, mean, sd float64
   for _, n := range p {
      sum += float64(n.GoldTimes[len(n.GoldTimes)-1])
   }
   mean = sum / float64(len(p))
   for _, n := range p {
      sd += math.Pow(float64(n.GoldTimes[len(n.GoldTimes)-1]) - mean, 2)
   }
   sd = math.Sqrt(sd / float64(len(p)))
   return sd
}

func TeamSum(ps []types.Player, s float64) float64 {
	var mean, sum, tSum float64
	for _, p := range ps {
		sum += float64(p.GoldTimes[len(p.GoldTimes)-1])
	}
	mean = sum / 5
	sd := networthStandardDeviation(ps)
	tSum = s * math.Pow((1 - sd / (mean + 1)), types.TeamSumExp)
	return tSum
}

func MacroCoeff(towers int, dur int) float64{
	standing := bits.OnesCount(uint(towers))
	destroyedWeight := math.Log(float64(12-standing) + 2.0)
    
    timeFactor := (1 - math.Pow(math.E, -float64(dur)/types.DurConst))
	return destroyedWeight * timeFactor
}

func TimeDom(dur int, adv []int, sign int) float64{
	var sum float64
	sum = 0
	for t, i := range adv {
		sum += math.Tanh(float64(sign * i)/types.AdvConst) * math.Pow(math.E, -float64(t)/types.DurConst)
	}
	sum /= float64(dur)
	return sum
}

func FaCoef(m *types.Match, p []types.Player, towers int, sign int) (float64, float64, float64, float64, []float64) {
	var coef float64
	sum, players := SumPlayers(p, m.Duration)
	tSum := TeamSum(m.Players, sum)
	macro := MacroCoeff(towers, m.Duration)
	dom := TimeDom(m.Duration, m.GoldAdv, sign)
	coef = tSum * macro * math.Pow(math.E, dom)
	return coef, tSum, macro, dom, players
}