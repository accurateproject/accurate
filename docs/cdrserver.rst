CDR Server
==========

An important component of every rating system is represented by the CDR Server. AccuRate includes an out of the box CDR Server component, controlable in the configuration file and supporting multiple interfaces for CDR feeds. This component makes the CDRs real-time accessible (influenced by the time of receiving them) to AccuRate subsystems.

Following interfaces are supported:


CDR-CGR
-------

Available as handler within http server.

To feed CDRs in via this interface, one must use url of the form: <http://$ip_configured:$port_configured/cdr_http>.

The CDR fields are received via http form (although for simplicity we support inserting them within query parameters as well) and are expected to be urlencoded in order to transport special characters reliably. All fields are expected by AccuRate as string, particular conversions being done on processing each CDR.
The fields received are split into two different categories based on AccuRate interest in them:

Primary fields: the fields which AccuRate needs for it's own operations and are stored into cdrs_primary table of storDb.

- tor: type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms>
- accid: represents the unique accounting id given by the telecom switch generating the CDR
- cdrhost: represents the IP address of the host generating the CDR (automatically populated by the server)
- cdrsource: formally identifies the source of the CDR (free form field)
- reqtype: matching the supported request types by the **AccuRate**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>.
- direction: matching the supported direction identifiers of the AccuRate <*out>
- tenant: tenant whom this record belongs
- category: free-form filter for this record, matching the category defined in rating profiles.
- account: account id (accounting subsystem) the record should be attached to
- subject: rating subject (rating subsystem) this record should be attached to
- destination: destination to be charged
- setup_time: set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
- answer_time: answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
- usage: event usage information (eg: in case of tor=*voice this will represent the total duration of a call)

Extra fields: any field coming in via the http request and not a member of primary fields list. These fields are stored as json encoded into *cdrs_extra* table of storDb.

Example of sample CDR generated simply using curl:
::

 curl --data "curl --data "tor=*voice&accid=iiaasbfdsaf&cdrhost=192.168.1.1&cdrsource=curl_cdr&reqtype=rated&direction=*out&tenant=192.168.56.66&category=call&account=dan&subject=dan&destination=%2B4986517174963&answer_time=1383813746&usage=1&sip_user=Jitsi&subject2=1003" http://127.0.0.1:2080/cdr_http


CDR-FS_JSON
-----------

Available as handler within http server, it implements the mechanism to store CDRs received from FreeSWITCH mod_json_cdr.

This interface is available at url:  <http://$ip_configured:$port_configured/freeswitch_json>.

This handler has a different implementation logic than the previous CDR-CGR, filtering fields received in the CDR from FreeSWITCH based on predefined configuration.
The mechanism of extracting CDR information out of JSON encoded CDR received from FreeSWITCH is the following:

- When receiving the CDR from FreeSWITCH, AccuRate will extract the content of ''variables'' object.
- Content of the ''variables'' will be filtered out and the following information will be stored into an internal CDR object:
   - Fields used by AccuRate in primary mediation, known as primary fields. These are:
      - uuid: internally generated uuid by FreeSWITCH for the call
      - sip_local_network_addr: IP address of the FreeSWITCH box generating the CDR
      - sip_call_id: call id out of SIP protocol
      - cgr_reqtype: request type as understood by the AccuRate
      - cgr_category: call category (optional)
      - cgr_tenant: tenant this call belongs to (optional)
      - cgr_account: account id in AccuRate (optional)
      - cgr_subject: rating subject in AccuRate (optional)
      - cgr_destination: destination being rated (optional)
      - user_name: username as seen by FreeSWITCH (considered if cgr_subject or cgr_account not present)
      - dialed_extension: destination number considered if cgr_destination is missing
   - Fields stored at request in cdr_extra and definable in configuration file under *extra_fields*.
- Once the content will be filtered, the real CDR object will be processed, stored into storDb under *cdrs_primary* and *cdrs_extra* tables and, if configured, it will be passed further for mediation.


CDR-RPC
-------

Available as RPC handler on top of CGR APIs exposed (in-process as well as GOB-RPC and JSON-RPC). This interface is used for example by CGR-SM component capturing the CDRs over event interface (eg: OpenSIPS or FreeSWITCH-ZeroConfig scenario)

The RPC function signature looks like this:
::

 CDRSV1.ProcessCdr(cdr *utils.StoredCdr, reply *string) error


The simplified StoredCdr object is represented by following:
::

 type StoredCdr struct {
   UniqueID          string
   OrderId        int64             // Stor order id used as export order id
   TOR            string            // type of record, meta-field, should map to one of the TORs hardcoded inside the server <*voice|*data|*sms>
   AccId          string            // represents the unique accounting id given by the telecom switch generating the CDR
   CdrHost        string            // represents the IP address of the host generating the CDR (automatically populated by the server)
   CdrSource      string            // formally identifies the source of the CDR (free form field)
   ReqType        string            // matching the supported request types by the **AccuRate**, accepted values are hardcoded in the server <prepaid|postpaid|pseudoprepaid|rated>.
   Direction      string            // matching the supported direction identifiers of the AccuRate <*out>
   Tenant         string            // tenant whom this record belongs
   Category       string            // free-form filter for this record, matching the category defined in rating profiles.
   Account        string            // account id (accounting subsystem) the record should be attached to
   Subject        string            // rating subject (rating subsystem) this record should be attached to
   Destination    string            // destination to be charged
   SetupTime      time.Time         // set-up time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
   AnswerTime     time.Time         // answer time of the event. Supported formats: datetime RFC3339 compatible, SQL datetime (eg: MySQL), unix timestamp.
   Usage          time.Duration     // event usage information (eg: in case of tor=*voice this will represent the total duration of a call)
   ExtraFields    map[string]string // Extra fields to be stored in CDR
}
