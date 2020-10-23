package calendar

import (
  "time"
  "sync"
  "github.com/google/uuid"
  "log"
  "io/ioutil"
  "gopkg.in/yaml.v2"
)

var AllCals Calendars

type HeaterCalendar struct {
  DoW byte `yaml:"DoW"`
  StartHour int `yaml:"StartHour"`
  EndHour int `yaml:"EndHour"`
  StartMinute int `yaml:"StartMinute"`
  EndMinute int `yaml:"EndMinute"`
  OnLow float32 `yaml:"OnLow"`
  OnHigh float32 `yaml:"OnHigh"`
  IsActive bool `yaml:"IsActive"`
}

func ( e *HeaterCalendar ) GetTemp ( tLow float32, tHigh float32 ) ( float32, float32 ) {
  var d byte
  cDate := time.Now ()
  d = 1 << int ( cDate.Weekday () )
  if d & e.DoW == 0 || e.IsActive == false {
    return tLow, tHigh
  }
  startDate := time.Date ( cDate.Year (), cDate.Month (), cDate.Day (), e.StartHour, e.StartMinute, 0, 0, cDate.Location () )
  endDate := time.Date ( cDate.Year (), cDate.Month (), cDate.Day (), e.EndHour, e.EndMinute, 0, 0, cDate.Location () )
  if cDate.After ( startDate ) && cDate.Before ( endDate ) {
    return e.OnLow, e.OnHigh
  }
  return tLow, tHigh
}

type Calendars struct {
  Cals map [ string ] HeaterCalendar
  m sync.Mutex
  fln string
}

func ( e *Calendars ) Init ( fln string ) {
  e.m.Lock ()
  defer e.m.Unlock ()
  e.Cals = make ( map [ string ] HeaterCalendar )
  e.fln = fln
  cfg, err := ioutil.ReadFile ( fln )
  if err != nil {
    log.Fatal ( "error reading calendar file" )
  }
  err = yaml.Unmarshal ( cfg, &e.Cals )
  if err != nil {
    log.Fatal ( "error parsing file" )
  }
}

func ( e *Calendars ) GetTemp ( tLow float32, tHigh float32 ) ( float32, float32 ) {
  rLow, rHigh := tLow, tHigh
  e.m.Lock ()
  defer e.m.Unlock ()
  for _,v := range e.Cals {
    rLow, rHigh = v.GetTemp ( tLow, tHigh )
  }
  return rLow, rHigh
}

func ( e *Calendars ) AddCalendar ( cid string, newCals HeaterCalendar ) string {
  u := uuid.New ()
  e.m.Lock ()
  defer e.m.Unlock ()
  var ret string
  if cid == "0" || cid == "" {
    e.Cals [ u.String () ] = newCals
    ret = u.String ()
  } else {
    e.Cals [ cid ] = newCals
    ret = cid
  }
  yy, _ := yaml.Marshal ( &e.Cals )
  _ = ioutil.WriteFile ( e.fln, yy, 0644 )
  return ret
}

func ( e *Calendars ) RemoveCalendar ( cid string ) {
  e.m.Lock ()
  defer e.m.Unlock ()
  delete ( e.Cals, cid )
  yy, _ := yaml.Marshal ( &e.Cals )
  _ = ioutil.WriteFile ( e.fln, yy, 0644 )
  return
}
