package engine

import (
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

var (
	testTPID     = "LoaderCSVTests"
	destinations = `
{"Tenant":"test", "Code": "49", "Tag": "GERMANY"}
{"Tenant":"test", "Code": "41", "Tag": "GERMANY_O2"}
{"Tenant":"test", "Code": "43", "Tag": "GERMANY_PREMIUM"}
{"Tenant":"test", "Code": "49", "Tag": "ALL"}
{"Tenant":"test", "Code": "41", "Tag": "ALL"}
{"Tenant":"test", "Code": "43", "Tag": "ALL"}
{"Tenant":"test", "Code": "0256", "Tag": "NAT"}
{"Tenant":"test", "Code": "0257", "Tag": "NAT"}
{"Tenant":"test", "Code": "0723", "Tag": "NAT"}
{"Tenant":"test", "Code": "+49", "Tag": "NAT"}
{"Tenant":"test", "Code": "0723", "Tag": "RET"}
{"Tenant":"test", "Code": "0724", "Tag": "RET"}
{"Tenant":"test", "Code": "0723045", "Tag": "SPEC"}
{"Tenant":"test", "Code": "+4971", "Tag": "PSTN_71"}
{"Tenant":"test", "Code": "+4972", "Tag": "PSTN_72"}
{"Tenant":"test", "Code": "+4970", "Tag": "PSTN_70"}
{"Tenant":"test", "Code": "447956", "Tag": "DST_UK_Mobile_BIG5"}
{"Tenant":"test", "Code": "112", "Tag": "URG"}
{"Tenant":"test", "Code": "444", "Tag": "EU_LANDLINE"}
{"Tenant":"test", "Code": "999", "Tag": "EXOTIC"}
`
	timings = `
{"Tenant":"test", "Tag": "WORKDAYS_00", "WeekDays": [1,2,3,4,5],"Time":"00:00:00"}
{"Tenant":"test", "Tag": "WORKDAYS_18", "WeekDays":[1,2,3,4,5], "Time":"18:00:00"}
{"Tenant":"test", "Tag": "WEEKENDS", "WeekDays":[6,0],"Time":"00:00:00"}
{"Tenant":"test", "Tag": "ONE_TIME_RUN", "Years":[2012], "Time":"*asap"}
`
	rates = `
{"Tenant":"test", "Tag":"R1", "Slots":[{"ConnectFee":0,"Rate":0.2,"RateUnit":"60s","RateIncrement":"1", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"R2", "Slots":[{"ConnectFee":0,"Rate":0.1,"RateUnit":"60","RateIncrement":"1", "GroupIntervalStart":"0"}]}
{"Tenant":"test", "Tag":"R3", "Slots":[{"ConnectFee":0,"Rate":0.05,"RateUnit":"60s","RateIncrement":"1s", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"R4", "Slots":[{"ConnectFee":1,"Rate":1,"RateUnit":"1s","RateIncrement":"1s", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"R5", "Slots":[{"ConnectFee":0,"Rate":0.5,"RateUnit":"1s","RateIncrement":"1s", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"LANDLINE_OFFPEAK", "Slots":[
        {"ConnectFee":0,"Rate":1,"RateUnit":"1s","RateIncrement":"60s", "GroupIntervalStart":"0s"},
        {"ConnectFee":0,"Rate":1,"RateUnit":"1s","RateIncrement":"1s", "GroupIntervalStart":"60s"}
]}
{"Tenant":"test", "Tag":"GBP_71", "Slots":[{"ConnectFee":0.000000,"Rate":5.55555,"RateUnit":"1s","RateIncrement":"1s", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"GBP_72", "Slots":[{"ConnectFee":0.000000,"Rate":7.77777,"RateUnit":"1","RateIncrement":"1", "GroupIntervalStart":"0"}]}
{"Tenant":"test", "Tag":"GBP_70", "Slots":[{"ConnectFee":0.000000,"Rate":1,"RateUnit":"1s","RateIncrement":"1s", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"RT_UK_Mobile_BIG5_PKG", "Slots":[{"ConnectFee":0.01,"Rate":0,"RateUnit":"20s","RateIncrement":"20s", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"RT_UK_Mobile_BIG5", "Slots":[{"ConnectFee":0.01,"Rate":0.10,"RateUnit":"1s","RateIncrement":"1s", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"R_URG", "Slots":[{"ConnectFee":0,"Rate":0,"RateUnit":"1","RateIncrement":"1", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"MX", "Slots":[{"ConnectFee":0,"Rate":1,"RateUnit":"1s","RateIncrement":"1s", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"DY", "Slots":[{"ConnectFee":0.15,"Rate":0.05,"RateUnit":"60s","RateIncrement":"1s", "GroupIntervalStart":"0s"}]}
{"Tenant":"test", "Tag":"CF", "Slots":[{"ConnectFee":1.12,"Rate":0,"RateUnit":"1s","RateIncrement":"1s", "GroupIntervalStart":"0s"}]}
`
	destinationRates = `
{"Tenant":"test", "Tag":"RT_STANDARD", "Bindings":[
          {"DestinationTag": "GERMANY", "RatesTag": "R1"},
          {"DestinationTag": "GERMANY_O2", "RatesTag": "R2"},
          {"DestinationTag": "GERMANY_PREMIUM", "RatesTag": "R2"}
]}
{"Tenant":"test", "Tag":"RT_DEFAULT", "Bindings":[{"DestinationTag": "ALL", "RatesTag": "R2"}]}
{"Tenant":"test", "Tag":"RT_STD_WEEKEND", "Bindings":[
          {"DestinationTag": "GERMANY", "RatesTag": "R2"},
          {"DestinationTag": "GERMANY_O2", "RatesTag": "R3"}
]}
{"Tenant":"test", "Tag":"P1", "Bindings":[{"DestinationTag": "NAT", "RatesTag": "R4"}]}
{"Tenant":"test", "Tag":"P2", "Bindings":[{"DestinationTag": "NAT", "RatesTag": "R5"}]}
{"Tenant":"test", "Tag":"T1", "Bindings":[{"DestinationTag": "NAT", "RatesTag": "LANDLINE_OFFPEAK"}]}
{"Tenant":"test", "Tag":"T2", "Bindings":[
          {"DestinationTag": "GERMANY", "RatesTag": "GBP_72"},
          {"DestinationTag": "GERMANY_O2", "RatesTag": "GBP_70"},
          {"DestinationTag": "GERMANY_PREMIUM", "RatesTag": "GBP_71"}
]}
{"Tenant":"test", "Tag":"GER", "Bindings":[{"DestinationTag": "GERMANY", "RatesTag": "R4"}]}
{"Tenant":"test", "Tag":"DR_UK_Mobile_BIG5_PKG", "Bindings":[{"DestinationTag": "DST_UK_Mobile_BIG5", "RatesTag": "RT_UK_Mobile_BIG5_PKG"}]}
{"Tenant":"test", "Tag":"DR_UK_Mobile_BIG5", "Bindings":[{"DestinationTag": "DST_UK_Mobile_BIG5", "RatesTag": "RT_UK_Mobile_BIG5"}]}
{"Tenant":"test", "Tag":"DATA_RATE", "Bindings":[{"DestinationTag": "*any", "RatesTag": "LANDLINE_OFFPEAK"}]}
{"Tenant":"test", "Tag":"RT_URG", "Bindings":[{"DestinationTag": "URG", "RatesTag": "R_URG"}]}
{"Tenant":"test", "Tag":"MX_FREE", "Bindings":[{"DestinationTag": "RET", "RatesTag": "MX", "MaxCost": 10,"MaxCostStrategy": "*free"}]}
{"Tenant":"test", "Tag":"MX_DISC", "Bindings":[{"DestinationTag": "RET", "RatesTag": "MX", "MaxCost": 10,"MaxCostStrategy": "*disconnect"}]}
{"Tenant":"test", "Tag":"RT_DY", "Bindings":[
          {"DestinationTag": "RET", "RatesTag": "DY", "RoundingMethod": "*up", "RoundingDecimals": 2},
          {"DestinationTag": "EU_LANDLINE", "RatesTag": "CF"}
]}
`
	ratingPlans = `
{"Tenant":"test", "Tag":"STANDARD", "Bindings":[
          {"DestinationRatesTag": "RT_STANDARD", "TimingTag": "WORKDAYS_00", "Weight": 10},
          {"DestinationRatesTag": "RT_STD_WEEKEND", "TimingTag": "WORKDAYS_18", "Weight": 10},
          {"DestinationRatesTag": "RT_STD_WEEKEND", "TimingTag": "WEEKENDS", "Weight": 10},
          {"DestinationRatesTag": "RT_URG", "TimingTag": "*any", "Weight": 20}
]}
{"Tenant":"test", "Tag":"PREMIUM", "Bindings":[
          {"DestinationRatesTag": "RT_STANDARD", "TimingTag": "WORKDAYS_00", "Weight": 10},
          {"DestinationRatesTag": "RT_STD_WEEKEND", "TimingTag": "WORKDAYS_18", "Weight": 10},
          {"DestinationRatesTag": "RT_STD_WEEKEND", "TimingTag": "WEEKENDS", "Weight": 10}
]}
{"Tenant":"test", "Tag":"DEFAULT", "Bindings":[
          {"DestinationRatesTag": "RT_DEFAULT", "TimingTag": "WORKDAYS_00", "Weight": 10},
          {"DestinationRatesTag": "RT_STD_WEEKEND", "TimingTag": "WEEKENDS", "Weight": 10}
]}
{"Tenant":"test", "Tag":"EVENING", "Bindings":[
          {"DestinationRatesTag": "P1", "TimingTag": "WORKDAYS_00", "Weight": 10},
          {"DestinationRatesTag": "P2", "TimingTag": "WORKDAYS_18", "Weight": 10},
          {"DestinationRatesTag": "P2", "TimingTag": "WEEKENDS", "Weight": 10}
]}
{"Tenant":"test", "Tag":"TDRT", "Bindings":[
          {"DestinationRatesTag": "T1", "TimingTag": "WORKDAYS_00", "Weight": 10},
          {"DestinationRatesTag": "T2", "TimingTag": "WEEKENDS", "Weight": 10}
]}
{"Tenant":"test", "Tag":"G", "Bindings":[
          {"DestinationRatesTag": "RT_STANDARD", "TimingTag": "WORKDAYS_00", "Weight": 10},
          {"DestinationRatesTag": "RT_STD_WEEKEND", "TimingTag": "WEEKENDS", "Weight": 10}
]}
{"Tenant":"test", "Tag":"R", "Bindings":[
          {"DestinationRatesTag": "P1", "TimingTag": "WORKDAYS_00", "Weight": 10},
          {"DestinationRatesTag": "P2", "TimingTag": "WEEKENDS", "Weight": 10}
]}
{"Tenant":"test", "Tag":"RP_UK_Mobile_BIG5_PKG", "Bindings":[
          {"DestinationRatesTag": "DR_UK_Mobile_BIG5_PKG", "TimingTag": "*any", "Weight": 10}
]}
{"Tenant":"test", "Tag":"RP_UK", "Bindings":[
          {"DestinationRatesTag": "DR_UK_Mobile_BIG5", "TimingTag": "*any", "Weight": 10}
]}
{"Tenant":"test", "Tag":"RP_DATA", "Bindings":[
          {"DestinationRatesTag": "DATA_RATE", "TimingTag": "*any", "Weight": 10}
]}
{"Tenant":"test", "Tag":"RP_MX", "Bindings":[
          {"DestinationRatesTag": "MX_DISC", "TimingTag": "WORKDAYS_00", "Weight": 10},
          {"DestinationRatesTag": "MX_FREE", "TimingTag": "WORKDAYS_18", "Weight": 10},
          {"DestinationRatesTag": "MX_FREE", "TimingTag": "WEEKENDS", "Weight": 10}
]}
{"Tenant":"test", "Tag":"GER_ONLY", "Bindings":[
          {"DestinationRatesTag": "GER", "TimingTag": "*any", "Weight": 10}
]}
{"Tenant":"test", "Tag":"ANY_PLAN", "Bindings":[
          {"DestinationRatesTag": "DATA_RATE", "TimingTag": "*any", "Weight": 10}
]}
{"Tenant":"test", "Tag":"DY_PLAN", "Bindings":[
          {"DestinationRatesTag": "RT_DY", "TimingTag": "*any", "Weight": 10}
]}
`
	ratingProfiles = `
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"rif:from:tm", "Activations":[
        {"ActivationTime":"2012-01-01T00:00:00Z", "RatingPlanTag":"PREMIUM", "FallbackSubjects":["danb"]},
        {"ActivationTime":"2012-02-28T00:00:00Z", "RatingPlanTag":"STANDARD", "FallbackSubjects":["danb"]}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"danb:87.139.12.167", "Activations":[
        {"ActivationTime":"2012-01-01T00:00:00Z", "RatingPlanTag":"STANDARD", "FallbackSubjects":["danb"]}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"danb", "Activations":[
        {"ActivationTime":"2012-01-01T00:00:00Z", "RatingPlanTag":"PREMIUM"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"rif", "Activations":[
        {"ActivationTime":"2012-01-01T00:00:00Z", "RatingPlanTag":"EVENING"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"rif", "Activations":[
        {"ActivationTime":"2012-02-28T00:00:00Z", "RatingPlanTag":"EVENING"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"dan", "Activations":[
        {"ActivationTime":"2012-01-01T00:00:00Z", "RatingPlanTag":"EVENING"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"minu", "Activations":[
        {"ActivationTime":"2012-01-01T00:00:00Z", "RatingPlanTag":"EVENING"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"*any", "Activations":[
        {"ActivationTime":"2012-01-01T00:00:00Z", "RatingPlanTag":"EVENING"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"one", "Activations":[
        {"ActivationTime":"2012-02-28T00:00:00Z", "RatingPlanTag":"STANDARD"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"inf", "Activations":[
        {"ActivationTime":"2012-02-28T00:00:00Z", "RatingPlanTag":"STANDARD", "FallbackSubjects":["inf"]}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"fall", "Activations":[
        {"ActivationTime":"2012-02-28T00:00:00Z", "RatingPlanTag":"PREMIUM", "FallbackSubjects":["rif"]}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"trp", "Activations":[
        {"ActivationTime":"2013-10-01T00:00:00Z", "RatingPlanTag":"TDRT", "FallbackSubjects":["rif", "danb"]}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"fallback1", "Activations":[
        {"ActivationTime":"2013-11-18T13:45:00Z", "RatingPlanTag":"G", "FallbackSubjects":["fallback2"]},
        {"ActivationTime":"2013-11-18T13:46:00Z", "RatingPlanTag":"G", "FallbackSubjects":["fallback2"]},
        {"ActivationTime":"2013-11-18T13:47:00Z", "RatingPlanTag":"G", "FallbackSubjects":["fallback2"]}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Subject":"fallback2", "Activations":[
        {"ActivationTime":"2013-11-18T13:45:00Z", "RatingPlanTag":"R", "FallbackSubjects":["rif"]}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"*any", "Activations":[
        {"ActivationTime":"2013-01-06T00:00:00Z", "RatingPlanTag":"RP_UK"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"discounted_minutes", "Activations":[
        {"ActivationTime":"2013-01-06T00:00:00Z", "RatingPlanTag":"RP_UK_Mobile_BIG5_PKG"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"data", "Subject":"rif", "Activations":[
        {"ActivationTime":"2013-01-06T00:00:00Z", "RatingPlanTag":"RP_DATA"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"max", "Activations":[
        {"ActivationTime":"2013-03-23T00:00:00Z", "RatingPlanTag":"RP_MX"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"nt", "Activations":[
        {"ActivationTime":"2012-02-28T00:00:00Z", "RatingPlanTag":"GER_ONLY"}
]}
{"Direction":"*in", "Tenant":"test", "Category":"LCR_STANDARD", "Subject":"max", "Activations":[
        {"ActivationTime":"2013-03-23T00:00:00Z", "RatingPlanTag":"RP_MX"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"money", "Activations":[
        {"ActivationTime":"2015-02-28T00:00:00Z", "RatingPlanTag":"EVENING"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"dy", "Activations":[
        {"ActivationTime":"2015-02-28T00:00:00Z", "RatingPlanTag":"DY_PLAN"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"block", "Activations":[
        {"ActivationTime":"2015-02-28T00:00:00Z", "RatingPlanTag":"DY_PLAN"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Subject":"round", "Activations":[
        {"ActivationTime":"2016-06-30T00:00:00Z", "RatingPlanTag":"DEFAULT"}
]}
`
	sharedGroups = `
{"Tenant":"test", "Tag":"SG1", "AccountParameters":{"*any":{"Strategy":"*lowest"}}}
{"Tenant":"test", "Tag":"SG2", "AccountParameters":{"*any":{"Strategy":"*lowest", "RatingSubject":"one"}}}
{"Tenant":"test", "Tag":"SG3", "AccountParameters":{"*any":{"Strategy":"*lowest"}}}
`

	lcrs = `
{"Direction":"*in", "Tenant": "test", "Category":"call", "Account":"*any", "Subject":"*any", "Activations":[
        {"ActivationTime":"2012-01-01T00:00:00Z", "Entries":[
             {"DestinationTag":"EU_LANDLINE", "RPCategory":"LCR_STANDARD", "Strategy":"*static","StrategyParams":"ivo;dan;rif", "Weight":10},
             {"DestinationTag":"*any", "RPCategory":"LCR_STANDARD", "Strategy":"*lowest_cost","StrategyParams":"", "Weight":20}
        ]}
]}
`
	actions = `
{"Tenant": "test", "Tag":"MINI", "Actions":[
         {"Action":"*topup_reset", "TOR":"*monetary", "Params":"{'Balance':{'Directions':'*out', 'Value':10, 'Weight':10}}", "Weight":10},
         {"Action":"*topup", "TOR":"*voice", "Params":"{'Balance':{'Directions':'*out', 'DestinationIDs':'NAT', 'RatingSubject':'test', 'Value':100, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"SHARED", "Actions":[
         {"Action":"*topup", "TOR":"*monetary", "Params":"{'Balance':{'Directions':'*out', 'Value':100, 'SharedGroups':'SG1', 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"TOPUP10_AC", "Actions":[
         {"Action":"*topup_reset", "TOR":"*monetary", "Filter":"{'ID':'b1'}", "Params":"{'Balance':{'ID':'b1', 'Directions':'*out', 'DestinationIDs':'*any', 'Value':1, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"TOPUP10_AC1", "Actions":[
         {"Action":"*topup_reset", "TOR":"*voice", "Params":"{'Balance':{'Directions':'*out', 'DestinationIDs':'DST_UK_Mobile_BIG5', 'RatingSubject':'discounted_minutes', 'Value':40, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"SE0", "Actions":[
         {"Action":"*topup_reset", "TOR":"*monetary", "Filter":"{'ID':'sg_bal'}", "Params":"{'Balance':{'ID':'sg_bal', 'Directions':'*out', 'SharedGroups':'SG2', 'Values':0, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"SE10", "Actions":[
         {"Action":"*topup_reset", "TOR":"*monetary", "Filter":"{'ID':'sg_bal'}", "Params":"{'Balance':{'ID':'sg_bal', 'Directions':'*out', 'SharedGroups':'SG2', 'Value':10, 'Weight':5}}", "Weight":10},
         {"Action":"*topup", "TOR":"*monetary", "Filter":"{'ID':'b1'}", "Params":"{'Balance':{'ID':'b1', 'Directions':'*out', 'Value':10, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"EE0", "Actions":[
         {"Action":"*topup_reset", "TOR":"*monetary", "Filter":"{'ID':'sg_bal'}", "Params":"{'Balance':{'ID':'sg_bal', 'Directions':'*out', 'SharedGroups':'SG3', 'Value':0, 'Weight':10}}", "Weight":10},
         {"Action":"*allow_negative", "TOR":"*monetary", "Filter":"{'ID':'b1'}", "Params":"{'Balance':{'ID':'b1', 'Directions':'*out', 'Value':0, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"DEFEE", "Actions":[
         {"Action":"*cdrlog", "Params":"{'CdrLogTemplate':{'Category':'^ddi','MediationRunId':'^did_run'}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"NEG", "Actions":[
         {"Action":"*allow_negative", "TOR":"*monetary", "Params":"{'Balance':{'Directions':'*out', 'Value':0, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"BLOCK", "Actions":[
         {"Action":"*topup", "TOR":"*monetary", "Filter":"{'ID':'bblocker'}", "Params":"{'Balance':{'ID':'bblocker', 'Directions':'*out','DestinationIDs':'NAT', 'Value':1, 'Weight':20,'Blocker':true}}", "Weight":20},
         {"Action":"*topup", "TOR":"*monetary", "Filter":"{'ID':'bfree'}", "Params":"{'Balance':{'ID':'bfree', 'Directions':'*out', 'Value':20, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"BLOCK_EMPTY", "Actions":[
         {"Action":"*topup", "TOR":"*monetary", "Filter":"{'ID':'bblocker'}", "Params":"{'Balance':{'ID':'bblocker', 'Directions':'*out','DestinationIDs':'NAT', 'Value':0, 'Weight':20,'Blocker':true}}", "Weight":20},
         {"Action":"*topup", "TOR":"*monetary", "Filter":"{'ID':'bfree'}", "Params":"{'Balance':{'ID':'bfree', 'Directions':'*out', 'Value':20, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"FILTER", "Actions":[
         {"Action":"*topup", "TOR":"*monetary","ExecFilter":"{'$and':[{'Value':{'$lt':0}},{'ID':{'$eq':'*default'}}]}", "Params":"{'Balance':{'ID':'bfree', 'Directions':'*out', 'Value':20, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"EXP", "Actions":[
         {"Action":"*topup", "TOR":"*voice", "Filter":"{'ID':'exp'}", "Params":"{'Balance':{'ID':'exp', 'ExpiryTime':'*monthly','TimingTags':['*any'], 'Directions':['*out'], 'Value':300, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"NOEXP", "Actions":[
         {"Action":"*topup", "TOR":"*voice", "Filter":"{'ID':'noexp'}", "Params":"{'Balance':{'ID':'noexp', 'ExpiryTime':'*unlimited','TimingTags':['*any'], 'Directions':['*out'], 'Value':50, 'Weight':10}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"VF", "Actions":[
         {"Action":"*debit", "TOR":"*monetary", "Params":"{'Balance':{'Directions':'*out', 'TimingTags':['*any'], 'Weight':10}, 'ValueFormula':{'Method':'*incremental','Args':{'Value':10, 'Interval':'month', 'Increment':'day'}}}", "Weight":10}
]}
{"Tenant": "test", "Tag":"LOG_WARNING", "Actions":[
         {"Action":"*log"}
]}
`
	actionPlans = `
{"Tenant":"test", "Tag":"MORE_MINUTES", "ActionTimings":[
         {"TimingTag":"ONE_TIME_RUN", "ActionsTag":"MINI", "Weight":10},
         {"TimingTag":"ONE_TIME_RUN", "ActionsTag":"SHARED", "Weight":10}
]}
{"Tenant":"test", "Tag":"TOPUP10_AT", "ActionTimings":[
         {"TimingTag":"*asap", "ActionsTag":"TOPUP10_AC", "Weight":10},
         {"TimingTag":"*asap", "ActionsTag":"TOPUP10_AC1", "Weight":10}
]}
{"Tenant":"test", "Tag":"TOPUP_SHARED0_AT", "ActionTimings":[
         {"TimingTag":"*asap", "ActionsTag":"SE0", "Weight":10}
]}
{"Tenant":"test", "Tag":"TOPUP_SHARED10_AT", "ActionTimings":[
         {"TimingTag":"*asap", "ActionsTag":"SE10", "Weight":10}
]}
{"Tenant":"test", "Tag":"TOPUP_EMPTY_AT", "ActionTimings":[
         {"TimingTag":"*asap", "ActionsTag":"EE0", "Weight":10}
]}
{"Tenant":"test", "Tag":"POST_AT", "ActionTimings":[
         {"TimingTag":"*asap", "ActionsTag":"NEG", "Weight":10}
]}
{"Tenant":"test", "Tag":"BLOCK_AT", "ActionTimings":[
         {"TimingTag":"*asap", "ActionsTag":"BLOCK", "Weight":10}
]}
{"Tenant":"test", "Tag":"BLOCK_EMPTY_AT", "ActionTimings":[
         {"TimingTag":"*asap", "ActionsTag":"BLOCK_EMPTY", "Weight":10}
]}
{"Tenant":"test", "Tag":"EXP_AT", "ActionTimings":[
         {"TimingTag":"*asap", "ActionsTag":"EXP", "Weight":10}
]}
`

	actionTriggers = `
{"Tenant":"test", "Tag":"STANDARD_TRIGGER", "Triggers":[
         {"UniqueID":"st0", "ThresholdType":"*min_event_counter", "ThresholdValue":10, "Recurent":false, "MinSleep":"0", "TOR":"*voice", "Filter":"{'Directions':{'$in':['*out']}, 'DestinationIDs':{'$in':['GERMANY_O2']}}", "ActionsTag":"SOME_1", "Weight":10},
         {"UniqueID":"st1", "ThresholdType":"*max_balance", "ThresholdValue":200, "Recurent":false, "MinSleep":"0", "TOR":"*voice", "Filter":"{'Directions':{'$in':['*out']}, 'DestinationIDs':{'$in':['GERMANY']}}", "ActionsTag":"SOME_2", "Weight":10}
]}
{"Tenant":"test", "Tag":"STANDARD_TRIGGERS", "Triggers":[
         {"ThresholdType":"*min_balance", "ThresholdValue":2, "Recurent":false, "MinSleep":"0", "TOR":"*monetary", "Filter":"{'Directions':{'$in':['*out']}}", "ActionsTag":"LOG_WARNING", "Weight":10},
         {"ThresholdType":"*max_balance", "ThresholdValue":20, "Recurent":false, "MinSleep":"0", "TOR":"*monetary", "Filter":"{'Directions':{'$in':['*out']}}", "ActionsTag":"LOG_WARNING", "Weight":10},
         {"ThresholdType":"*max_event_counter", "ThresholdValue":5, "Recurent":false, "MinSleep":"0", "TOR":"*monetary", "Filter":"{'Directions':{'$in':['*out']}}, 'DestinationIDs':{'$in':['FS_USERS']}", "ActionsTag":"LOG_WARNING", "Weight":10}
]}
{"Tenant":"test", "Tag":"CDRST1_WARN_ASR", "Triggers":[
         {"ThresholdType":"*min_asr", "ThresholdValue":45, "Recurent":true, "MinSleep":"1h", "MinQueuedItems":3, "ActionsTag":"CDRST_WARN_HTTP", "Weight":10}
]}
{"Tenant":"test", "Tag":"CDRST1_WARN_ACD", "Triggers":[
         {"ThresholdType":"*min_acd", "ThresholdValue":10, "Recurent":true, "MinSleep":"1h", "MinQueuedItems":5, "ActionsTag":"CDRST_WARN_HTTP", "Weight":10}
]}
{"Tenant":"test", "Tag":"CDRST1_WARN_ACC", "Triggers":[
         {"ThresholdType":"*max_acc", "ThresholdValue":10, "Recurent":true, "MinSleep":"0", "MinQueuedItems":5, "ActionsTag":"CDRST_WARN_HTTP", "Weight":10}
]}
{"Tenant":"test", "Tag":"CDRST2_WARN_ASR", "Triggers":[
         {"ThresholdType":"*min_asr", "ThresholdValue":30, "Recurent":true, "MinSleep":"0", "MinQueuedItems":5, "ActionsTag":"CDRST_WARN_HTTP", "Weight":10}
]}
{"Tenant":"test", "Tag":"CDRST2_WARN_ACD", "Triggers":[
         {"ThresholdType":"*min_acd", "ThresholdValue":3, "Recurent":true, "MinSleep":"0", "MinQueuedItems":5, "ActionsTag":"CDRST_WARN_HTTP", "Weight":10}
]}
`
	accountActions = `
{"Tenant":"test", "Account":"minitsboy", "ActionPlanTags":["MORE_MINUTES"], "ActionTriggerTags":["STANDARD_TRIGGER"]}
{"Tenant":"test", "Account":"12345", "ActionPlanTags":["TOPUP10_AT"], "ActionTriggerTags":["STANDARD_TRIGGERS"]}
{"Tenant":"test", "Account":"123456", "ActionPlanTags":["TOPUP10_AT"], "ActionTriggerTags":["STANDARD_TRIGGERS"]}
{"Tenant":"test", "Account":"dy", "ActionPlanTags":["TOPUP10_AT"], "ActionTriggerTags":["STANDARD_TRIGGERS"]}
{"Tenant":"test", "Account":"remo", "ActionPlanTags":["TOPUP10_AT"]}
{"Tenant":"test", "Account":"empty0", "ActionPlanTags":["TOPUP_SHARED0_AT"]}
{"Tenant":"test", "Account":"empty10", "ActionPlanTags":["TOPUP_SHARED10_AT"]}
{"Tenant":"test", "Account":"emptyX", "ActionPlanTags":["TOPUP_EMPTY_AT"]}
{"Tenant":"test", "Account":"emptyY", "ActionPlanTags":["TOPUP_EMPTY_AT"]}
{"Tenant":"test", "Account":"post", "ActionPlanTags":["POST_AT"]}
{"Tenant":"test", "Account":"alodis", "ActionPlanTags":["TOPUP_EMPTY_AT"], "AllowNegative":true, "Disabled":true}
{"Tenant":"test", "Account":"block", "ActionPlanTags":["BLOCK_AT"]}
{"Tenant":"test", "Account":"block_empty", "ActionPlanTags":["BLOCK_EMPTY_AT"]}
{"Tenant":"test", "Account":"expo", "ActionPlanTags":["EXP_AT"]}
{"Tenant":"test", "Account":"expnoexp"}
{"Tenant":"test", "Account":"vf"}
{"Tenant":"test", "Account":"round", "ActionPlanTags":["TOPUP10_AT"]}
`

	derivedChargers = `
{"Direction":"*out", "Tenant":"test", "Category":"call", "Account":"dan", "Subject":"dan", "Chargers":[
         {"RunID":"extra1", "RunFilters":"", "Fields":"{'RequestType':{'$set':'*prepaid'}, 'Account':{'$set':'rif'}, 'Subject':{'$set':'rif'}}"},
         {"RunID":"extra2", "Fields":"{'Account':{'$set':'ivo'}, 'Subject':{'$set':'ivo'}}"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Account":"dan", "Subject":"*any", "Chargers":[
         {"RunID":"extra1", "Fields":"{'Account':{'$set':'rif2'}, 'Subject':{'$set':'rif2'}}"}
]}

`
	cdrStats = `
{"Tenant":"test", "Tag":"CDRST1", "QueueLength":5, "TimeWindow":"60m", "Metrics":["ASR", "ACD", "ACC"], "Filter":"{'SetupInterval':{'$btw':['2014-07-29T15:00:00Z','2014-07-29T16:00:00Z']}, 'TOR':'*voice', 'CdrHost':'87.139.12.167', 'CdrSource':'FS_JSON', 'ReqType':'*rated', 'Direction':'*out', 'Tenant':'test', 'Category':'call', 'Account':'dan', 'Subject':'dan', 'DestinationPrefix':'49', 'PddInterval':{'$btw':['3m','7m']}, 'UsageInterval':{'$btw':['5m','10m']}, 'Supplier':'suppl1', 'DisconnectCause':'NORMAL_CLEARING', 'MediationRunIds':'default', 'RatedAccount':'rif', 'RatedSubject':'rif', 'CostInterval':{'$btw':[0,2]}}",  "ActionTriggerTags":["STANDARD_TRIGGERS"]}
{"Tenant":"test", "Tag":"CDRST2", "QueueLength":10, "TimeWindow":"10m", "Metrics":["ASR", "ACD"], "Filter":"{'Tenant':'test', 'Category':'call'}"}
`
	users = `
{"Tenant":"test", "Name":"rif", "Masked":false, "Weight":10, "Query":"{'test0':{'$usr':'val0'}, 'test1':{'$usr': 'val1'}}"}
{"Tenant":"test", "Name":"dan", "Masked":false, "Weight":10, "Query":"{'another':{'$usr':'value'}}"}
{"Tenant":"test", "Name":"mas", "Masked":true, "Weight":10, "Query":"{'another':{'$usr':'value'}}"}
{"Tenant":"t1", "Name":"t1", "Weight":10,
 "Query":"{'sip_from_host':{'$in':['206.222.29.2','206.222.29.3','206.222.29.4','206.222.29.5','206.222.29.6']}, 'Destination':{'$crepl':['^9023(\\\\d+)','${1}']}, 'Account':{'$usr':'t1'}, 'direction':{'$usr': 'outbound'}, 'Subject':{'$repl':['^9023(\\\\d+)','${1}']}}"}
`
	aliases = `
{"Direction":"*out", "Tenant":"test", "Category":"call", "Account":"dan", "Subject":"dan", "Context":"*rating", "Values":[
        {"DestinationTag":"EU_LANDLINE", "Weight":10, "Fields":"{'Subject':{'$crepl':['(dan|rif)','${1}1']}, 'Cli':{'$rpl':['0723','0724']}}"},
        {"DestinationTag":"GLOBAL1", "Weight":20, "Fields":"{'Subject':{'$rpl':['dan','dan2']}}"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"*any", "Account":"*any", "Subject":"*any", "Context":"*rating", "Values":[
        {"DestinationTag":"*any", "Weight":20, "Fields":"{'Subject':{'$rpl':['*any','rif1']}}"},
        {"DestinationTag":"*any", "Weight":10, "Fields":"{'Subject':'dan1'}"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"0", "Account":"a1", "Subject":"a1", "Context":"*rating", "Values":[
        {"DestinationTag":"*any", "Weight":10, "Fields":"{'Subject':{'$rpl':['a1','minu']}, 'Account':{'$rpl':['a1','minu']}}"}
]}
{"Direction":"*out", "Tenant":"test", "Category":"call", "Account":"remo", "Subject":"remo", "Context":"*rating", "Values":[
        {"DestinationTag":"", "Weight":10, "Fields":"{'Subject':{'$rpl':['remo','minu']}, 'Account':{'$rpl':['remo','rif1']}}"}
]}
`

	resLimits = `
{"Tenant":"test", "Tag":"ResGroup1", "Filters":[
        {"Type": "*string", "FieldName":"Account", "Values":["1001","1002"], "ActivationTime":"2014-07-29T15:00:00Z", "Weight":10, "Limit":2, "ActionTriggersTag":""},
        {"Type": "*string_prefix", "FieldName":"Destination", "Values":["10","20"], "ActivationTime":"2014-07-29T15:00:00Z", "Weight":10},
        {"Type": "*cdr_stats", "FieldName":"", "Values":["CDRST1:*min_ASR:34","CDRST_1001:*min_ASR:20"]},
        {"Type": "*rsr_fields", "FieldName":"", "Values":["Subject(~^1.*1$)","Destination(1002)"], "ActivationTime":"", "Weight":10}
]}
{"Tenant":"test", "Tag":"ResGroup2", "Filters":[
        {"Type": "*destinations", "FieldName":"Destination", "Values":["DST_FS"], "ActivationTime":"2014-07-29T15:00:00Z", "Weight":10, "Limit":2}
]}
`
)

var csvr *TpReader

func loadFromJSON() { // called bt TestMain
	tpr := NewTpReader(ratingStorage, accountingStorage, *config.Get().General.DefaultTimezone)

	if err := utils.LoadJSON(strings.NewReader(destinations), func() interface{} { return &utils.TpDestination{} }, tpr.LoadDestination); err != nil {
		log.Printf("error loading destinations: %v", err)
	}

	if err := utils.LoadJSON(strings.NewReader(timings), func() interface{} { return &utils.TpTiming{} }, tpr.LoadTiming); err != nil {
		log.Printf("error loading timings: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(rates), func() interface{} { return &utils.TpRate{} }, tpr.LoadRate); err != nil {
		log.Printf("error loading rates: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(destinationRates), func() interface{} { return &utils.TpDestinationRate{} }, tpr.LoadDestinationRate); err != nil {
		log.Printf("error loading destinationRates: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(ratingPlans), func() interface{} { return &utils.TpRatingPlan{} }, tpr.LoadRatingPlan); err != nil {
		log.Printf("error loading ratingPlans: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(ratingProfiles), func() interface{} { return &utils.TpRatingProfile{} }, tpr.LoadRatingProfile); err != nil {
		log.Printf("error loading ratingProfiles: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(sharedGroups), func() interface{} { return &utils.TpSharedGroup{} }, tpr.LoadSharedGroup); err != nil {
		log.Printf("error loading sharedGroups: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(lcrs), func() interface{} { return &utils.TpLcrRule{} }, tpr.LoadLCR); err != nil {
		log.Printf("error loading lcrs: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(actions), func() interface{} { return &utils.TpActionGroup{} }, tpr.LoadActionGroup); err != nil {
		log.Printf("error loading actions: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(actionPlans), func() interface{} { return &utils.TpActionPlan{} }, tpr.LoadActionPlan); err != nil {
		log.Printf("error loading actionPlans: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(actionTriggers), func() interface{} { return &utils.TpActionTrigger{} }, tpr.LoadActionTrigger); err != nil {
		log.Printf("error loading actionTriggers: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(accountActions), func() interface{} { return &utils.TpAccountAction{} }, tpr.LoadAccountAction); err != nil {
		log.Printf("error loading accountActions: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(derivedChargers), func() interface{} { return &utils.TpDerivedCharger{} }, tpr.LoadDerivedCharger); err != nil {
		log.Printf("error loading derivedChargers: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(cdrStats), func() interface{} { return &utils.TpCdrStats{} }, tpr.LoadCdrStats); err != nil {
		log.Printf("error loading cdrStats: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(users), func() interface{} { return &utils.TpUser{} }, tpr.LoadUser); err != nil {
		log.Printf("error loading users: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(aliases), func() interface{} { return &utils.TpAlias{} }, tpr.LoadAlias); err != nil {
		log.Printf("error loading aliases: %v", err)
	}
	if err := utils.LoadJSON(strings.NewReader(resLimits), func() interface{} { return &utils.TpResourceLimit{} }, tpr.LoadResourceLimit); err != nil {
		log.Printf("error loading resLimits: %v", err)
	}
	cache2go.Flush("test")
	ratingStorage.PreloadRatingCache()
	accountingStorage.PreloadAccountingCache()
}

func TestLoadDestinations(t *testing.T) {
	count, err := ratingStorage.Count(ColDst)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 27 {
		t.Error("failed to load destinations: ", count)
	}

	dests, err := ratingStorage.GetDestinations("test", "", "NAT", utils.DestExact, utils.CACHED)
	if err != nil {
		t.Fatalf("error getting destinations: %v", err)
	}
	if len(dests) != 4 {
		t.Error("failed getting NAT: ", utils.ToIJSON(dests))
	}
	dests, err = ratingStorage.GetDestinations("test", "0723", "", utils.DestExact, utils.CACHED)
	if err != nil {
		t.Fatalf("error getting destinations: %v", err)
	}
	if len(dests) != 3 {
		t.Error("failed getting 0723: ", utils.ToIJSON(dests))
	}
}

func TestLoadTimings(t *testing.T) {
	count, err := ratingStorage.Count(ColTmg)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 4 {
		t.Error("failed to load timings: ", count)
	}

	timing, err := ratingStorage.GetTiming("test", "WORKDAYS_00")
	if err != nil {
		t.Fatalf("error getting timing: %v", err)
	}
	if !reflect.DeepEqual(timing, &Timing{
		Tenant:    "test",
		Name:      "WORKDAYS_00",
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		Time:      "00:00:00",
	}) {
		t.Error("Error loading timing: ", timing)
	}
	timing, err = ratingStorage.GetTiming("test", "WORKDAYS_18")
	if err != nil {
		t.Fatalf("error getting timing: %v", err)
	}
	if !reflect.DeepEqual(timing, &Timing{
		Tenant:    "test",
		Name:      "WORKDAYS_18",
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
		Time:      "18:00:00",
	}) {
		t.Error("Error loading timing: ", timing)
	}
	timing, err = ratingStorage.GetTiming("test", "WEEKENDS")
	if err != nil {
		t.Fatalf("error getting timing: %v", err)
	}
	if !reflect.DeepEqual(timing, &Timing{
		Tenant:    "test",
		Name:      "WEEKENDS",
		Years:     utils.Years{},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{time.Saturday, time.Sunday},
		Time:      "00:00:00",
	}) {
		t.Error("Error loading timing: ", timing)
	}
	timing, err = ratingStorage.GetTiming("test", "ONE_TIME_RUN")
	if err != nil {
		t.Fatalf("error getting timing: %v", err)
	}
	if !reflect.DeepEqual(timing, &Timing{
		Tenant:    "test",
		Name:      "ONE_TIME_RUN",
		Years:     utils.Years{2012},
		Months:    utils.Months{},
		MonthDays: utils.MonthDays{},
		WeekDays:  utils.WeekDays{},
		Time:      "*asap",
	}) {
		t.Error("Error loading timing: ", timing)
	}
}

func TestLoadRates(t *testing.T) {
	count, err := ratingStorage.Count(ColRts)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 15 {
		t.Error("failed to load rates: ", count)
	}

	rate, err := ratingStorage.GetRate("test", "R1")
	if err != nil {
		t.Fatalf("error getting rate: %v", err)
	}
	expctRs := &RateSlot{0, 0.2, 60 * time.Second, 1 * time.Second, 0}
	rateSlot := rate.Slots[0]
	if err != nil {
		t.Error("Error loading rate: ", rate, err.Error())
	} else if !reflect.DeepEqual(rateSlot, expctRs) {
		t.Error("Error loading rateSlot: ", rateSlot, expctRs)
	}
	rate, err = ratingStorage.GetRate("test", "R2")
	if err != nil {
		t.Fatalf("error getting rate: %v", err)
	}
	rateSlot = rate.Slots[0]
	expctRs = &RateSlot{0, 0.1, 60 * time.Second, 1 * time.Second, 0}
	if !reflect.DeepEqual(rateSlot, expctRs) {
		t.Error("Error loading rateSlot: ", rateSlot)
	}
	rate, err = ratingStorage.GetRate("test", "R3")
	if err != nil {
		t.Fatalf("error getting rate: %v", err)
	}
	rateSlot = rate.Slots[0]
	expctRs = &RateSlot{0, 0.05, 60 * time.Second, 1 * time.Second, 0}
	if !reflect.DeepEqual(rateSlot, expctRs) {
		t.Error("Error loading rateSlot: ", rateSlot)
	}
	rate, err = ratingStorage.GetRate("test", "R4")
	if err != nil {
		t.Fatalf("error getting rate: %v", err)
	}
	rateSlot = rate.Slots[0]
	expctRs = &RateSlot{1, 1.0, 1 * time.Second, 1 * time.Second, 0}
	if !reflect.DeepEqual(rateSlot, expctRs) {
		t.Error("Error loading rateSlot: ", rateSlot)
	}
	rate, err = ratingStorage.GetRate("test", "R5")
	if err != nil {
		t.Fatalf("error getting rate: %v", err)
	}
	rateSlot = rate.Slots[0]
	expctRs = &RateSlot{0, 0.5, 1 * time.Second, 1 * time.Second, 0}
	if !reflect.DeepEqual(rateSlot, expctRs) {
		t.Error("Error loading rateSlot: ", rateSlot)
	}
	rate, err = ratingStorage.GetRate("test", "LANDLINE_OFFPEAK")
	if err != nil {
		t.Fatalf("error getting rate: %v", err)
	}
	rateSlot = rate.Slots[0]
	expctRs = &RateSlot{0, 1, 1 * time.Second, 60 * time.Second, 0}
	if !reflect.DeepEqual(rateSlot, expctRs) {
		t.Error("Error loading rateSlot: ", rateSlot)
	}
	rate, err = ratingStorage.GetRate("test", "LANDLINE_OFFPEAK")
	if err != nil {
		t.Fatalf("error getting rate: %v", err)
	}
	rateSlot = rate.Slots[1]
	expctRs = &RateSlot{0, 1, 1 * time.Second, 1 * time.Second, 60 * time.Second}
	if !reflect.DeepEqual(rateSlot, expctRs) {
		t.Error("Error loading rateSlot: ", rateSlot)
	}
}

func TestLoadDestinationRates(t *testing.T) {
	count, err := ratingStorage.Count(ColDrt)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 15 {
		t.Error("failed to load destination rates: ", count)
	}

	drs, err := ratingStorage.GetDestinationRate("test", "RT_STANDARD")
	if err != nil {
		t.Fatalf("error getting detination rate: %v", err)
	}
	dr := &DestinationRate{
		Tenant: "test",
		Name:   "RT_STANDARD",
		Bindings: map[string]*DestinationRateBinding{
			"49_R1": &DestinationRateBinding{
				DestinationCode: "49",
				DestinationName: "GERMANY",
				RateID:          "R1",
				MaxCost:         0,
				MaxCostStrategy: "",
			},
			"41_R2": &DestinationRateBinding{
				DestinationCode: "41",
				DestinationName: "GERMANY_O2",
				RateID:          "R2",
			},
			"43_R2": &DestinationRateBinding{
				DestinationCode: "43",
				DestinationName: "GERMANY_PREMIUM",
				RateID:          "R2",
			},
		},
	}
	if !reflect.DeepEqual(drs, dr) {
		t.Errorf("Error loading destination rate: \n%s \n%s", utils.ToIJSON(drs.Bindings), utils.ToIJSON(dr.Bindings))
	}
	drs, err = ratingStorage.GetDestinationRate("test", "RT_DEFAULT")
	if err != nil {
		t.Fatalf("error getting destination rate: %v", err)
	}
	if !reflect.DeepEqual(drs, &DestinationRate{
		Tenant: "test",
		Name:   "RT_DEFAULT",
		Bindings: map[string]*DestinationRateBinding{
			"49_R2": &DestinationRateBinding{
				DestinationCode: "49",
				DestinationName: "ALL",
				RateID:          "R2",
			},
			"41_R2": &DestinationRateBinding{
				DestinationCode: "41",
				DestinationName: "ALL",
				RateID:          "R2",
			},
			"43_R2": &DestinationRateBinding{
				DestinationCode: "43",
				DestinationName: "ALL",
				RateID:          "R2",
			},
		},
	}) {
		t.Errorf("Error loading destination rate: %s", utils.ToIJSON(drs.Bindings))
	}
	drs, err = ratingStorage.GetDestinationRate("test", "RT_STD_WEEKEND")
	if err != nil {
		t.Fatalf("error getting destination rate: %v", err)
	}
	if !reflect.DeepEqual(drs, &DestinationRate{
		Tenant: "test",
		Name:   "RT_STD_WEEKEND",
		Bindings: map[string]*DestinationRateBinding{
			"49_R2": &DestinationRateBinding{
				DestinationCode: "49",
				DestinationName: "GERMANY",
				RateID:          "R2",
			},
			"41_R3": &DestinationRateBinding{
				DestinationCode: "41",
				DestinationName: "GERMANY_O2",
				RateID:          "R3",
			},
		},
	}) {
		t.Error("Error loading destination rate: ", utils.ToIJSON(drs))
	}
	drs, err = ratingStorage.GetDestinationRate("test", "P1")
	if err != nil {
		t.Fatalf("error getting destination rate: %v", err)
	}
	if !reflect.DeepEqual(drs, &DestinationRate{
		Tenant: "test",
		Name:   "P1",
		Bindings: map[string]*DestinationRateBinding{
			"0257_R4": &DestinationRateBinding{
				DestinationCode: "0257",
				DestinationName: "NAT",
				RateID:          "R4",
			},
			"0723_R4": &DestinationRateBinding{
				DestinationCode: "0723",
				DestinationName: "NAT",
				RateID:          "R4",
			},
			"+49_R4": &DestinationRateBinding{
				DestinationCode: "+49",
				DestinationName: "NAT",
				RateID:          "R4",
			},
			"0256_R4": &DestinationRateBinding{
				DestinationCode: "0256",
				DestinationName: "NAT",
				RateID:          "R4",
			},
		},
	}) {
		t.Error("Error loading destination rate: ", utils.ToIJSON(drs))
	}
	drs, err = ratingStorage.GetDestinationRate("test", "P2")
	if err != nil {
		t.Fatalf("error getting destination rate: %v", err)
	}
	if !reflect.DeepEqual(drs, &DestinationRate{
		Tenant: "test",
		Name:   "P2",
		Bindings: map[string]*DestinationRateBinding{
			"0256_R5": &DestinationRateBinding{
				DestinationCode: "0256",
				DestinationName: "NAT",
				RateID:          "R5",
			},
			"0257_R5": &DestinationRateBinding{
				DestinationCode: "0257",
				DestinationName: "NAT",
				RateID:          "R5",
			},
			"0723_R5": &DestinationRateBinding{
				DestinationCode: "0723",
				DestinationName: "NAT",
				RateID:          "R5",
			},
			"+49_R5": &DestinationRateBinding{
				DestinationCode: "+49",
				DestinationName: "NAT",
				RateID:          "R5",
			},
		},
	}) {
		t.Error("Error loading destination rate: ", utils.ToIJSON(drs))
	}
	drs, err = ratingStorage.GetDestinationRate("test", "T1")
	if err != nil {
		t.Fatalf("error getting destination rate: %v", err)
	}
	if !reflect.DeepEqual(drs, &DestinationRate{
		Tenant: "test",
		Name:   "T1",
		Bindings: map[string]*DestinationRateBinding{
			"0256_LANDLINE_OFFPEAK": &DestinationRateBinding{
				DestinationCode: "0256",
				DestinationName: "NAT",
				RateID:          "LANDLINE_OFFPEAK",
			},
			"0257_LANDLINE_OFFPEAK": &DestinationRateBinding{
				DestinationCode: "0257",
				DestinationName: "NAT",
				RateID:          "LANDLINE_OFFPEAK",
			},
			"0723_LANDLINE_OFFPEAK": &DestinationRateBinding{
				DestinationCode: "0723",
				DestinationName: "NAT",
				RateID:          "LANDLINE_OFFPEAK",
			},
			"+49_LANDLINE_OFFPEAK": &DestinationRateBinding{
				DestinationCode: "+49",
				DestinationName: "NAT",
				RateID:          "LANDLINE_OFFPEAK",
			},
		},
	}) {
		t.Error("Error loading destination rate: ", utils.ToIJSON(drs))
	}
	drs, err = ratingStorage.GetDestinationRate("test", "T2")
	if err != nil {
		t.Fatalf("error getting destination rate: %v", err)
	}
	if !reflect.DeepEqual(drs, &DestinationRate{
		Tenant: "test",
		Name:   "T2",
		Bindings: map[string]*DestinationRateBinding{
			"49_GBP_72": &DestinationRateBinding{
				DestinationCode: "49",
				DestinationName: "GERMANY",
				RateID:          "GBP_72",
			},
			"41_GBP_70": &DestinationRateBinding{
				DestinationCode: "41",
				DestinationName: "GERMANY_O2",
				RateID:          "GBP_70",
			},
			"43_GBP_71": &DestinationRateBinding{
				DestinationCode: "43",
				DestinationName: "GERMANY_PREMIUM",
				RateID:          "GBP_71",
			},
		},
	}) {
		t.Error("Error loading destination destination rate: ", utils.ToIJSON(drs))
	}
}

func TestLoadRatingPlans(t *testing.T) {
	count, err := ratingStorage.Count(ColRpl)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 14 {
		t.Error("failed to load rating plans: ", count)
	}

	rplan, err := ratingStorage.GetRatingPlan("test", "STANDARD", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting rating plan: %v", err)
	}
	expected := &RatingPlan{
		Tenant: "test",
		Name:   "STANDARD",
		Timings: map[string]*RITiming{
			"4c954a": &RITiming{
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
			},
			"4d5932": &RITiming{
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "18:00:00",
			},
			"a60bfb": &RITiming{
				WeekDays:  utils.WeekDays{time.Saturday, time.Sunday},
				StartTime: "00:00:00",
			},
			"30eab3": &RITiming{
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"970150": &RIRate{
				ConnectFee: dec.NewFloat(0),
				MaxCost:    dec.New(),
				Rates: RateGroups{
					&RateInfo{
						GroupIntervalStart: 0,
						Value:              dec.NewFloat(0.2),
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
			},
			"2d0cfa": &RIRate{
				ConnectFee: dec.NewFloat(0),
				MaxCost:    dec.New(),
				Rates: RateGroups{
					&RateInfo{
						GroupIntervalStart: 0,
						Value:              dec.NewFloat(0.1),
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
			},
			"387308": &RIRate{
				ConnectFee: dec.NewFloat(0),
				MaxCost:    dec.New(),
				Rates: RateGroups{
					&RateInfo{
						GroupIntervalStart: 0,
						Value:              dec.NewFloat(0.05),
						RateIncrement:      time.Second,
						RateUnit:           time.Minute,
					},
				},
			},
			"3bcc60": &RIRate{
				ConnectFee: dec.NewFloat(0),
				MaxCost:    dec.New(),
				Rates: RateGroups{
					&RateInfo{
						GroupIntervalStart: 0,
						Value:              dec.NewFloat(0),
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
			},
		},
		DRates: map[string]*DRate{
			"df3e96": &DRate{
				Timing: "4c954a",
				Rating: "2d0cfa",
				Weight: 10,
			},
			"0fcffe": &DRate{
				Timing: "4c954a",
				Rating: "970150",
				Weight: 10,
			},
			"498aa6": &DRate{
				Timing: "4d5932",
				Rating: "2d0cfa",
				Weight: 10,
			},
			"1e3c4d": &DRate{
				Timing: "4d5932",
				Rating: "387308",
				Weight: 10,
			},
			"c0ece7": &DRate{
				Timing: "a60bfb",
				Rating: "387308",
				Weight: 10,
			},
			"7ec088": &DRate{
				Timing: "a60bfb",
				Rating: "2d0cfa",
				Weight: 10,
			},
			"49c1f1": &DRate{
				Timing: "30eab3",
				Rating: "3bcc60",
				Weight: 20,
			},
		},
		DestinationRates: map[string]*DRateHelper{
			"41": &DRateHelper{
				DRateKeys: map[string]struct{}{
					"df3e96": struct{}{},
					"1e3c4d": struct{}{},
					"c0ece7": struct{}{},
				},
				CodeName: "GERMANY_O2",
			},
			"43": &DRateHelper{
				DRateKeys: map[string]struct{}{
					"df3e96": struct{}{},
				},
				CodeName: "GERMANY_PREMIUM",
			},
			"49": &DRateHelper{
				DRateKeys: map[string]struct{}{
					"7ec088": struct{}{},
					"498aa6": struct{}{},
					"0fcffe": struct{}{},
				},
				CodeName: "GERMANY",
			},
			"112": &DRateHelper{
				DRateKeys: map[string]struct{}{
					"49c1f1": struct{}{},
				},
				CodeName: "URG",
			},
		},
	}
	if !reflect.DeepEqual(rplan, expected) {
		t.Errorf("Received: %s", utils.ToIJSON(rplan))
	}

	anyTiming := &RITiming{
		StartTime:  "00:00:00",
		EndTime:    "",
		cronString: "",
		//tag:        utils.ANY,
	}

	rplan, err = ratingStorage.GetRatingPlan("test", "ANY_PLAN", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting rating plan: %v", err)
	}
	if !reflect.DeepEqual(rplan.Timings["30eab3"], anyTiming) {
		t.Errorf("Error using *any timing in rating plans: %+v : %+v", rplan.Timings["30eab3"], anyTiming)
	}
}

func TestLoadRatingProfiles(t *testing.T) {
	count, err := ratingStorage.Count(ColRpf)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 24 {
		t.Error("failed to load rating profiles: ", count)
	}

	rp, err := ratingStorage.GetRatingProfile(utils.OUT, "test", "0", "trp", false, utils.CACHED)
	if err != nil {
		t.Fatalf("error getting rating profile: %v", err)
	}
	expected := &RatingProfile{
		Direction: utils.OUT,
		Tenant:    "test",
		Category:  "0",
		Subject:   "trp",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2013, 10, 1, 3, 0, 0, 0, time.Local),
			RatingPlanID:    "TDRT",
			FallbackKeys:    []string{"rif", "danb"},
			CdrStatQueueIDs: []string{},
		}},
	}
	if !reflect.DeepEqual(rp, expected) {
		t.Errorf("Error loading rating profile: %+v", utils.ToIJSON(rp.RatingPlanActivations))
	}
}

func TestLoadActions(t *testing.T) {
	count, err := ratingStorage.Count(ColAct)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 18 {
		t.Error("failed to load actions: ", count)
	}

	as1, err := ratingStorage.GetActionGroup("test", "MINI", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting action: %v", err)
	}

	expected := &ActionGroup{
		Tenant: "test",
		Name:   "MINI",
		Actions: Actions{&Action{
			ActionType: TOPUP_RESET,
			TOR:        utils.MONETARY,
			Params:     `{"Balance":{"Directions":"*out", "Value":10, "Weight":10}}`,
			Weight:     10,
		},
			&Action{
				ActionType: TOPUP,
				TOR:        utils.VOICE,
				Params:     `{"Balance":{"Directions":"*out", "DestinationIDs":"NAT", "RatingSubject":"test", "Value":100, "Weight":10}}`,
				Weight:     10,
			},
		},
	}
	if !reflect.DeepEqual(as1, expected) {
		t.Errorf("Error loading action1: %s", utils.ToIJSON(as1))
	}
	as2, err := ratingStorage.GetActionGroup("test", "SHARED", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting action: %v", err)
	}
	expected = &ActionGroup{
		Tenant: "test",
		Name:   "SHARED",
		Actions: Actions{
			&Action{
				ActionType: TOPUP,
				Weight:     10,
				TOR:        utils.MONETARY,
				Params:     `{"Balance":{"Directions":"*out", "Value":100, "SharedGroups":"SG1", "Weight":10}}`,
			},
		},
	}
	if !reflect.DeepEqual(as2, expected) {
		t.Errorf("Error loading action: %s", utils.ToIJSON(as2))
	}
	as3, err := ratingStorage.GetActionGroup("test", "DEFEE", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting action: %v", err)
	}
	expected = &ActionGroup{
		Tenant: "test",
		Name:   "DEFEE",
		Actions: Actions{
			&Action{
				ActionType: CDRLOG,
				Weight:     10,
				Params:     `{"CdrLogTemplate":{"Category":"^ddi","MediationRunId":"^did_run"}}`,
			},
		},
	}
	if !reflect.DeepEqual(as3.Actions[0], expected.Actions[0]) {
		t.Errorf("Error loading action: %s", utils.ToIJSON(as3))
	}
}

func TestLoadSharedGroups(t *testing.T) {
	count, err := ratingStorage.Count(ColShg)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 4 {
		t.Error("failed to load shared groups: ", count)
	}

	sg1, err := ratingStorage.GetSharedGroup("test", "SG1", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting shared group: %v", err)
	}
	expected := &SharedGroup{
		Tenant: "test",
		Name:   "SG1",
		AccountParameters: map[string]*SharingParam{
			"*any": &SharingParam{
				Strategy:      "*lowest",
				RatingSubject: "",
			},
		},
		MemberIDs: utils.StringMap{},
	}
	if !reflect.DeepEqual(sg1, expected) {
		t.Errorf("Error loading shared group: %+v, expected %+v", sg1.MemberIDs, expected.MemberIDs)
	}
	sg2, err := ratingStorage.GetSharedGroup("test", "SG2", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting shared group: %v", err)
	}
	expected = &SharedGroup{
		Tenant: "test",
		Name:   "SG2",
		AccountParameters: map[string]*SharingParam{
			"*any": &SharingParam{
				Strategy:      "*lowest",
				RatingSubject: "one",
			},
		},
		MemberIDs: utils.StringMap{
			"empty0":  true,
			"empty10": true,
		},
	}
	if !reflect.DeepEqual(sg2, expected) {
		t.Error("Error loading shared group: ", utils.ToIJSON(sg2))
	}
}

func TestLoadLCRs(t *testing.T) {
	count, err := ratingStorage.Count(ColLcr)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 1 {
		t.Error("failed to load lcrs: ", count)
	}

	lcr, err := ratingStorage.GetLCR(utils.IN, "test", "call", "*any", "*any", false, utils.CACHED)
	if err != nil {
		t.Fatalf("error getting lcr: %v", err)
	}
	expected := &LCR{
		Direction: utils.IN,
		Tenant:    "test",
		Category:  "call",
		Account:   "*any",
		Subject:   "*any",
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2012, 1, 1, 2, 0, 0, 0, time.Local),
				Entries: []*LCREntry{
					&LCREntry{
						DestinationID:  "EU_LANDLINE",
						RPCategory:     "LCR_STANDARD",
						Strategy:       "*static",
						StrategyParams: "ivo;dan;rif",
						Weight:         10,
					},
					&LCREntry{
						DestinationID:  "*any",
						RPCategory:     "LCR_STANDARD",
						Strategy:       "*lowest_cost",
						StrategyParams: "",
						Weight:         20,
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(lcr, expected) {
		t.Errorf("Error loading lcr %s: ", utils.ToIJSON(lcr))
	}
}

func TestLoadActionPlans(t *testing.T) {
	count, err := ratingStorage.Count(ColApl)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 9 {
		t.Error("failed to load action plans: ", count)
	}

	apl, err := ratingStorage.GetActionPlan("test", "MORE_MINUTES", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting lcr: %v", err)
	}
	expected := &ActionPlan{
		Tenant: "test",
		Name:   "MORE_MINUTES",
		ActionTimings: []*ActionTiming{
			&ActionTiming{
				UUID: apl.ActionTimings[0].UUID,
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "MINI",
			},
			&ActionTiming{
				UUID: apl.ActionTimings[1].UUID,
				Timing: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{2012},
						StartTime: utils.ASAP,
					},
				},
				Weight:    10,
				ActionsID: "SHARED",
			},
		},
	}
	if !reflect.DeepEqual(apl, expected) {
		t.Errorf("Error loading action timing:\n%s", utils.ToIJSON(apl))
	}
}

func TestLoadActionTriggers(t *testing.T) {
	count, err := ratingStorage.Count(ColApl)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 9 {
		t.Error("failed to load action triggers: ", count)
	}

	atg, err := ratingStorage.GetActionTriggers("test", "STANDARD_TRIGGER", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting action trigger: %v", err)
	}
	atr := atg.ActionTriggers[0]
	atr.parentGroup = nil
	expected := &ActionTrigger{
		UniqueID:       "st0",
		ThresholdType:  utils.TRIGGER_MIN_EVENT_COUNTER,
		ThresholdValue: dec.NewFloat(10),
		TOR:            utils.VOICE,
		Filter:         `{"Directions":{"$in":["*out"]}, "DestinationIDs":{"$in":["GERMANY_O2"]}}`,
		Weight:         10,
		ActionsID:      "SOME_1",
	}
	if !reflect.DeepEqual(atr, expected) {
		t.Errorf("Error loading action trigger: %+v", atr)
	}
	atr = atg.ActionTriggers[1]
	atr.parentGroup = nil
	expected = &ActionTrigger{
		UniqueID:       "st1",
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: dec.NewFloat(200),
		TOR:            utils.VOICE,
		Filter:         `{"Directions":{"$in":["*out"]}, "DestinationIDs":{"$in":["GERMANY"]}}`,
		Weight:         10,
		ActionsID:      "SOME_2",
	}
	if !reflect.DeepEqual(atr, expected) {
		t.Errorf("Error loading action trigger: %s", utils.ToIJSON(atr))
	}
}

func TestLoadAccountActions(t *testing.T) {
	count, err := accountingStorage.Count(ColAcc)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 36 {
		t.Error("failed to load accounts: ", count)
	}

	aa, err := accountingStorage.GetAccount("test", "minitsboy")
	if err != nil {
		t.Fatalf("error getting account: %v", err)
	}

	expected := &Account{
		Tenant: "test",
		Name:   "minitsboy",
		UnitCounters: UnitCounters{
			utils.VOICE: []*UnitCounter{
				&UnitCounter{
					CounterType: "*event",
					Counters: CounterFilters{
						&CounterFilter{
							Value:  dec.NewFloat(0),
							Filter: ``,
						},
					},
				},
			},
		},
		//ActionTriggers: csvr.actionsTriggers["STANDARD_TRIGGER"],
	}
	_ = aa
	_ = expected
	/*
		// set propper uuid
		for i, atr := range aa.ActionTriggers {
			csvr.actionsTriggers["STANDARD_TRIGGER"][i].ID = atr.ID
		}
		for i, b := range aa.UnitCounters[utils.VOICE][0].Counters {
			expected.UnitCounters[utils.VOICE][0].Counters[i].Filter.ID = b.Filter.ID
		}
		if !reflect.DeepEqual(aa.UnitCounters[utils.VOICE][0].Counters[0], expected.UnitCounters[utils.VOICE][0].Counters[0]) {
			t.Errorf("Error loading account action: %+v", utils.ToIJSON(aa.UnitCounters[utils.VOICE][0].Counters[0].Filter))
		}
		// test that it does not overwrite balances
		existing, err := accountingStorage.GetAccount(aa.ID)
		if err != nil || len(existing.BalanceMap) != 2 {
			t.Errorf("The account was not set before load: %+v", existing)
		}
		accountingStorage.SetAccount(aa)
		existing, err = accountingStorage.GetAccount(aa.ID)
		if err != nil || len(existing.BalanceMap) != 2 {
			t.Errorf("The set account altered the balances: %+v", existing)
		}*/

	//TODO: check action plan binding
}

func TestLoadDerivedChargers(t *testing.T) {
	count, err := ratingStorage.Count(ColDcs)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 2 {
		t.Error("failed to load derived chargers: ", count)
	}

	chargers1, err := ratingStorage.GetDerivedChargers("*out", "test", "call", "dan", "dan", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting account: %v", err)
	}
	expCharger1 := &utils.DerivedChargerGroup{
		Direction:      "*out",
		Tenant:         "test",
		Category:       "call",
		Account:        "dan",
		Subject:        "dan",
		DestinationIDs: nil,
		Chargers: []*utils.DerivedCharger{
			&utils.DerivedCharger{
				RunID:     "extra1",
				RunFilter: ``,
				Fields:    `{"RequestType":{"$set":"*prepaid"}, "Account":{"$set":"rif"}, "Subject":{"$set":"rif"}}`,
			},
			&utils.DerivedCharger{
				RunID:     "extra2",
				RunFilter: ``,
				Fields:    `{"Account":{"$set":"ivo"}, "Subject":{"$set":"ivo"}}`,
			},
		},
	}
	expCharger2 := &utils.DerivedChargerGroup{
		Direction:      "*out",
		Tenant:         "test",
		Category:       "call",
		Account:        "dan",
		Subject:        "*any",
		DestinationIDs: nil,
		Chargers: []*utils.DerivedCharger{
			&utils.DerivedCharger{
				RunID:     "extra1",
				RunFilter: ``,
				Fields:    `{"Account":{"$set":"rif2"}, "Subject":{"$set":"rif2"}}`,
			},
		},
	}

	if len(chargers1) != 2 || !chargers1[0].Equal(expCharger1) || !chargers1[1].Equal(expCharger2) {
		t.Errorf("received: %s", utils.ToIJSON(chargers1))
	}
}

func TestLoadCdrStats(t *testing.T) {
	count, err := ratingStorage.Count(ColCrs)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 2 {
		t.Error("failed to load cdr stats: ", count)
	}

	cs1, err := ratingStorage.GetCdrStats("test", "CDRST1")
	if err != nil {
		t.Fatalf("error getting cdr stat: %v", err)
	}

	cdrStats1 := &CdrStats{
		Tenant:      "test",
		Name:        "CDRST1",
		QueueLength: 5,
		TimeWindow:  60 * time.Minute,
		Metrics:     []string{"ASR", "ACD", "ACC"},
		Filter:      `{"SetupInterval":{"$btw":["2014-07-29T15:00:00Z","2014-07-29T16:00:00Z"]}, "TOR":"*voice", "CdrHost":"87.139.12.167", "CdrSource":"FS_JSON", "ReqType":"*rated", "Direction":"*out", "Tenant":"test", "Category":"call", "Account":"dan", "Subject":"dan", "DestinationPrefix":"49", "PddInterval":{"$btw":["3m","7m"]}, "UsageInterval":{"$btw":["5m","10m"]}, "Supplier":"suppl1", "DisconnectCause":"NORMAL_CLEARING", "MediationRunIds":"default", "RatedAccount":"rif", "RatedSubject":"rif", "CostInterval":{"$btw":[0,2]}}`,
		TriggerIDs:  utils.StringMap{"STANDARD_TRIGGERS": true},
	}
	if !reflect.DeepEqual(cs1, cdrStats1) {
		t.Errorf("Unexpected stats %s", utils.ToIJSON(cs1))
	}
}

func TestLoadUsers(t *testing.T) {
	count, err := accountingStorage.Count(ColUsr)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 4 {
		t.Error("failed to load users: ", count)
	}

	u1, err := accountingStorage.GetUser("test", "rif")
	if err != nil {
		t.Fatalf("error getting user: %v", err)
	}
	user1 := &UserProfile{
		Tenant: "test",
		Name:   "rif",
		Index:  make(map[string]string),
		Query:  `{"test0":{"$usr":"val0"}, "test1":{"$usr": "val1"}}`,
		Weight: 10,
	}

	if !reflect.DeepEqual(u1, user1) {
		t.Errorf("expected %s got %s", utils.ToIJSON(user1), utils.ToIJSON(u1))
	}

	u2, err := accountingStorage.GetUser("t1", "t1")
	if err != nil {
		t.Fatalf("error getting user: %v", err)
	}
	user2 := &UserProfile{
		Tenant: "t1",
		Name:   "t1",
		Index:  make(map[string]string),
		Query:  `{"sip_from_host":{"$in":["206.222.29.2","206.222.29.3","206.222.29.4","206.222.29.5","206.222.29.6"]}, "Destination":{"$crepl":["^9023(\\d+)","${1}"]}, "Account":{"$usr":"t1"}, "direction":{"$usr": "outbound"}, "Subject":{"$repl":["^9023(\\d+)","${1}"]}}`,
		Weight: 10,
	}
	if !reflect.DeepEqual(u2, user2) {
		t.Errorf("expected %s got %s", utils.ToIJSON(user2), utils.ToIJSON(u2))
	}
}

func TestLoadAliases(t *testing.T) {
	count, err := accountingStorage.Count(ColAls)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 5 {
		t.Error("failed to aliases: ", count)
	}

	a1, err := accountingStorage.GetAlias("*out", "test", "call", "dan", "dan", "*rating", utils.CACHED)
	if err != nil {
		t.Fatalf("error getting alias: %v", err)
	}
	alias1 := &Alias{
		Direction: "*out",
		Tenant:    "test",
		Category:  "call",
		Account:   "dan",
		Subject:   "dan",
		Context:   "*rating",
		Values: AliasValues{
			&AliasValue{
				DestinationID: "GLOBAL1",
				Fields:        `{"Subject":{"$rpl":["dan","dan2"]}}`,
				Weight:        20,
			},
			&AliasValue{
				DestinationID: "EU_LANDLINE",
				Fields:        `{"Subject":{"$crepl":["(dan|rif)","${1}1"]}, "Cli":{"$rpl":["0723","0724"]}}`,
				Weight:        10,
			},
		},
	}

	a1.Values[0].fields = nil
	a1.Values[1].fields = nil
	if !reflect.DeepEqual(a1.Values, alias1.Values) {
		t.Errorf("Unexpected alias %s", utils.ToIJSON(a1))
	}
}

func TestLoadResourceLimits(t *testing.T) {
	count, err := accountingStorage.Count(ColRL)
	if err != nil {
		t.Fatalf("error getting count: %v", err)
	}
	if count != 0 {
		t.Error("failed to res limits: ", count)
	}

	rl, err := accountingStorage.GetResourceLimit("", false, utils.CACHED)
	if err != nil {
		t.Fatalf("error getting res limit: %v", err)
	}
	at, _ := utils.ParseTimeDetectLayout("2014-07-29T15:00:00Z", *config.Get().General.DefaultTimezone)
	eResLimits := map[string]*ResourceLimit{
		"ResGroup1": &ResourceLimit{
			ID: "ResGroup1",
			Filters: []*RequestFilter{
				&RequestFilter{Type: MetaString, FieldName: "Account", Values: []string{"1001", "1002"}},
				&RequestFilter{Type: MetaStringPrefix, FieldName: "Destination", Values: []string{"10", "20"}},
				&RequestFilter{Type: MetaCDRStats, Values: []string{"CDRST1:*min_ASR:34", "CDRST_1001:*min_ASR:20"}},
				&RequestFilter{Type: MetaRSRFields, Values: []string{"Subject(~^1.*1$)", "Destination(1002)"}},
			},
			ActivationTime: at,
			Weight:         10,
			Limit:          2,
		},
		"ResGroup2": &ResourceLimit{
			ID: "ResGroup2",
			Filters: []*RequestFilter{
				&RequestFilter{Type: MetaDestinations, FieldName: "Destination", Values: []string{"DST_FS"}},
			},
			ActivationTime: at,
			Weight:         10,
			Limit:          2,
		},
	}
	if count == len(eResLimits) {
		t.Error("Failed to load resourcelimits: ", count)
	}
	if reflect.DeepEqual(eResLimits, rl) {
		t.Errorf("Expecting: %+v, received: %+v", eResLimits, rl)
	}

}
