Visualization Tutorial
=================

This tutorial will explain how to get spirit-box logs visualized in Grafana. All dependancies are installed locally on the machine for this tutorial.

Prerequisites: 

+ Python3
+ Pip3
+ Telegraf 1.23.2
+ Elasticsearch 7.16.3
+ Grafana 9.0.5

## Scrape log data using Telegraf

The first step is changing the telegraf configuration file that will parse the Json in the logs. The default conf file is located:

```/etc/telegraf/telegraf.conf```

Replace the default conf file with the file provided.

Replace `...` in ```files = ["..."]``` with the filepath to the spirit-box log.

If running elasticsearch in a container, replace ```urls = ["..."]``` in outputs.elasticsearch with the url of the cluster. (**not tested**)

Once the conf file is finished, you may then run telegraf. By default it is located at `/usr/bin/telegraf`. Confirm that the log was successfully imported by visiting ```http://localhost:9200/_cat/indices?v```

![image](https://user-images.githubusercontent.com/56091505/186519577-e4bab209-594e-4da4-bb4c-b8ca5aea7c6d.png)

A new index will be created. Remember this for later.

## Configure Elasticsearch datasource

In Grafana, select `Data sources` from the settings side bar.

![image](https://user-images.githubusercontent.com/56091505/186520273-0ea5030c-40d8-4c22-9c5c-5bfb14224650.png)

Select `Add data source` and then scroll down and select `Elasticsearch`. 

Now we will configure the data source.

+ Choose an appropriate **Name** for the data source. Remember this for later
+ **URL** to elasticsearch (http://localhost:9200 by default)
+ **Index name** which is the index created previously, ex. telegraf-2022.08.23
+ **ElasticSearch version** 7.10+

Then select to `Save & test`.

![image](https://user-images.githubusercontent.com/56091505/186523415-99721b61-7620-4d61-86df-c5219e369f45.png)

Messages will appear confirming that the index and time field are ok, and that the data source is updated.

## Generate dashboard

Now that the log is imported into Elasticsearch and the datasource is configured in Grafana, we will generate the dashboard.

Install Jinja2 with the command `pip3 install Jinja2`.

Run the dashboard generator script `python3 generate_dashboard.py <data_source>`.

The template.json must be in the same directory as the script. data_source is the name of the data source chosen in **Configure Elasticsearch datasource**.

Confirm that no errors are reported by the script and the new dashboard has been added to Grafana.

Q & A
=================

Q: How do I get Grafana running?
+ Open a terminal and navigate to `/usr/share/grafana` . Run `grafana-server web` .

Q: Why are panels showing duplicate events?
+ This is an issue with telegraf. You must quit telegraf as soon as stdout shows the data. There may be a way to set telegraf's configuration to end after one import. Otherwise it will duplicate records into elasticsearch.

Q: Does this work with Elasticsearch versions over 8.0?
+ Grafana support for Elasticsearch 8.0+ is experimental so it may not work. To be safe use Elasticsearch 7.16.3.

Q: How do I delete Elasticsearch indices?
+ `curl -X DELETE "localhost:9200/<index>?pretty"`

Q: Does this work with InfluxDB?
+ Yes. Add InfluxDB as an output in the telegraf configuration file. Then set InfluxDB as a datasource in Grafana. The dashboard generator script however will not work.
