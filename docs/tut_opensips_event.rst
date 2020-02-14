OpenSIPS_ interaction via  *event_datagram*
===========================================

Scenario
--------

- OpenSIPS out of *residential* configuration generated. 

 - Considering the following users (with configs hardcoded in the *opensips.cfg* configuration script): 1002-postpaid, 1003-pseudoprepaid, 1004-rated, 1007-rated.
 - For simplicity we configure no authentication (WARNING: Not for production usage).

- **AccuRate** with following components:

 - CGR-SM started as translator between OpenSIPS_ and **cgr-rater** for both authorization events (pseudoprepaid) as well as CDR ones.
 - CGR-CDRS component processing raw CDRs from CGR-SM component and storing them inside CGR StorDB.
 - CGR-CDRE exporting rated CDRs from CGR StorDB (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path browsable with git client at */tmp/cgr_history*).


Starting OpenSIPS_ with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/osips_async/opensips/etc/init.d/opensips start

To verify that OpenSIPS_ is running we run the console command:

::

 opensipsctl moni


Starting **AccuRate** with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/osips_async/cgrates/etc/init.d/cgrates start

Make sure that cgrates is running

::

 cgr-console status


CDR processing
--------------

At the end of each call OpenSIPS_ will generate an CDR event and due to automatic handler registration built in **AccuRate-SM** component, this will be directed towards the port configured inside *cgrates.json*. This event will reach inside **AccuRate** through the *SM* component (close to real-time). Once in-there it will be instantly rated and be ready for export. 


**AccuRate** Usage
-----------------

Since it is common to most of the tutorials, the example for **AccuRate** usage is provided in a separate page `here <http://cgrates.readthedocs.org/en/latest/tut_cgrates_usage.html>`_


.. _OpenSIPS: http://www.opensips.org/
