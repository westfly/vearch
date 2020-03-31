# -*- coding: UTF-8 -*-

import logging
import pytest
import requests
import json

logging.basicConfig()
logger = logging.getLogger(__name__)

__author__ = 'wangjiangjuan' + 'jiazijian'
__date__ = '2019-07-22 09:25:00'
__description__ = """ """

ip = "127.0.0.1"
ip_db = ip + ":8817"
ip_data = ip + ":9001"
db_name = "ts_db"
space_name_mmap = "ts_space"
space_name_rocksdb = "vector_space_rocksdb"
fileData = "/home/vearch/test/data/test_data.json"
# fileData = "D:\\tool\\VDB\\vectorbase\\test\\data\\test_data.json"

# query_list = [
#     '{"query": {"filter": [{"range": {"int": {"gte": 0, "lte": 9999}}}], "sum": [{"field": "vector", "feature": [0.9405091], "format": "normalization"}]}, "size": 10, "quick": false, "vector_value": false}',
#     '{"query": {"sum": [{"field": "vector", "feature": [0.9405091], "symbol": ">=", "value": 0.9}], "filter": [{"range": {"int": {"gte": 1, "lte": 1000}}}, {"term": {"string_tags": ["28", "2", "29"], "operator": "or"}}]}, "size": 10, "quick": false, "vector_value": false}',
# ]
query_resultList = []

@pytest.mark.author('')
@pytest.mark.level(2)
@pytest.mark.cover(["VDB"])

def test_stats():
    logging.info("_cluster_information")
    url = "http://" + ip_db + "/_cluster/stats"
    response = requests.get(url)
    print("cluster_stats:" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"status\":200")>=0

def test_health():
    url = "http://" + ip_db + "/_cluster/health"
    response = requests.get(url)
    print("cluster_health---\n" + response.text)
    assert response.status_code == 200
    # assert response.text.find("\"status\":\"green\"")>=0

def test_server():
    url = "http://" + ip_db + "/list/server"
    response = requests.get(url)
    print("list_server---\n" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"msg\":\"success\"")>=0

logger.info("database")
def test_dblist():
    url = "http://" + ip_db + "/list/db"
    response = requests.get(url)
    print("list_db---\n" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"msg\":\"success\"")>=0

def test_createDB():
    logger.info("------------")
    url = "http://" + ip_db + "/db/_create"
    headers = {"content-type": "application/json"}
    data = {
        'name':db_name
    }
    response = requests.put(url, headers=headers, data=json.dumps(data))
    print("db_create---\n" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"msg\":\"success\"")>=0

def test_dbsearch():
    # url = "http://" + ip_db + "/db/" + db_name
    url = "http://" + ip_db + "/db/ts_db"
    response = requests.get(url)
    print("db_search---\n" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"msg\":\"success\"")>=0

def test_listspace():
    url = "http://" + ip_db + "/list/space?db=" + db_name
    response = requests.get(url)
    print("list_space---\n" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"msg\":\"success\"")>=0

logger.info("space")
def test_createSpaceMmap():
    url = "http://" + ip_db + "/space/" + db_name +"/_create"
    headers = {"content-type": "application/json"}
    data = {
        "name": "ts_space",
        "dynamic_schema": "strict",
        "partition_num": 1,
        "replica_num": 1,
        "engine": {
            "name": "gamma",
            "index_size": 10000,
            "max_size": 10000*10000
        },
        "properties": {
            "string": {
                "type": "keyword",
                "index": True
            },
            "int": {
                "type": "integer",
                "index": True
            },
            "float": {
                "type": "float",
                "index": True
            },
            "vector": {
                "type": "vector",
                "model_id": "img",
                "dimension": 128,
                # "store_type": store_type,                  #默认 "Mmap"
                # "store_param": {"cache_size": cache_size}, #默认 max_size*dimension*typeof(float)
                "format": "normalization"
            },
            "string_tags": {
                "type": "string",
                "array": True,
                "index": True
            },
            "int_tags": {
                "type": "integer",
                "array": True,
                "index": True
            },
            "float_tags": {
                "type": "float",
                "array": True,
                "index": True
            }
        },
        "models": [{
            "model_id": "vgg16",
            "fields": ["string"],
            "out": "feature"
        }]
    }
    # data = {
    #     "name": "ts_space",
    #     "dynamic_schema": "strict",
    #     "partition_num": 3,
    #     "replica_num": 3,
    #     "engine": {"name": "gamma", "index_size": 10000, "max_size": 20000000},
    #     "properties": {
    #         "sku": {
    #             "type": "integer",
    #             "index": "false"
    #         },
    #         "img_url": {
    #             "type": "keyword",
    #             "index": "false"
    #         },
    #         "cid1": {
    #             "type": "integer",
    #             "index": "true"
    #         },
    #         "cid2": {
    #             "type": "integer",
    #             "index": "true"
    #         },
    #         "cid3": {
    #             "type": "integer",
    #             "index": "true"
    #         },
    #         "spu": {
    #             "type": "integer",
    #             "index": "false"
    #         },
    #         "brand_id": {
    #             "type": "integer",
    #             "index": "false"
    #         },
    #         "feature": {
    #             "type": "vector",
    #             "model_id": "img",
    #             "dimension": 512,
    #             # "retrieval_type": "GPU",
    #             "store_param": {"cache_size":40960}
    #         }
    #     },
    #     "models": [{
    #         "model_id": "vgg16",
    #         "fields": ["url"],
    #         "out": "feature"
    #     }]
    # }
    print(url+"---"+json.dumps(data))
    response = requests.put(url, headers=headers, data=json.dumps(data))
    print("space_create---\n" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"msg\":\"success\"")>=0

def test_getspace():
    url = "http://" + ip_db + "/space/"+db_name+"/" + space_name_mmap
    response = requests.get(url)
    print("get_space---\n" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"msg\":\"success\"")>=0


# def test_changemember():
#     url = "http://" + ip_db + "/partition/change_member"
#     headers = {"content-type": "application/json"}
#     data = {
#         "partition_id":7,
#         "node_id":1,
#         "method":0
#     }
#     response = requests.post(url, headers=headers, data=json.dumps(data))
#     print("change_member:" + response.text)
#     assert response.status_code == 200
#     assert response.text.find("\"msg\":\"success\"")>=0

logger.info("router(PS)")
def test_insertWithId():
    logger.info("insert")
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            flag = 0
            flag1 = idStr.split(':')[1].replace('\"','')
            id = str(int(flag1)+flag)
            data = "{"+dataLine.split(',', 1)[1]
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_mmap + "/" + id
            response = requests.post(url, headers=headers, data=data)
            print("insertWithID:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"status\":201")>=0

def test_searchById():
    logger.info("test_searchById")
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            # id = eval(idStr.split(':')[1])
            flag = 0
            flag1 = idStr.split(':')[1].replace('\"','')
            id = str(int(flag1)+flag)
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_mmap + "/" + id
            response = requests.get(url)
            print("searchById:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"found\":true")>=0

def test_insterNoId():
    logger.info("insertDataNoId")
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            data = "{"+dataLine.split(',', 1)[1]
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_mmap
            response = requests.post(url, headers=headers, data=data)
            print("insertNoID:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"successful\":1")>=0

def test_searchByFeature():
    headers = {"content-type": "application/json"}
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_mmap+"/_search?size=100"
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0")>=0

def test_searchByFeatureandFilter():
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_mmap+"/_search"
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            string_tags = feature["string_tags"]
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "filter": [{
                        "string_tags": string_tags
                    }],
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0") >= 0

def test_searchByFeatureandRange():
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_mmap+"/_search"
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(1),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            string_tags = feature["string_tags"]
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "filter": [{
                        "range": {
                            "int" : {
                                "gte" : 0,
                                "lte" : 0
                            }
                        }
                    }],
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            # print("data:" + json.dumps(data))
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0") >= 0

def test_searchByTerm():
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_mmap+"/_search"
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(1),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            string_tags = feature["string_tags"]
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "filter": [{
                        "term": {
                            "string": "0AW1mK_j19FyJvn5NR4Eb",
                            "operator": "or"
                        }
                    }],
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            # data = {
            #     "query": {
            #         "filter": [{
            #             "term": {
            #                 "string": "0AW1mK_j19FyJvn5NR4Eb"
            #             }
            #         }]
            #     }
            # }
            # print("data:" + json.dumps(data))
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0") >= 0

# def test_searchAll():
#     url = "http://" + ip_data + "/"+db_name+"/"+space_name_mmap+"/_search"
#     headers = {"content-type": "application/json"}
#     with open(fileData, "r") as dataLine1:
#         for querydata in query_list:
#             print("querydata:" + querydata)
#             for i, dataLine in zip(range(1),dataLine1):
#                 idStr = dataLine.split(',', 1)[0].replace('{', '')
#                 id = eval(idStr.split(':')[1])
#                 feature = "{"+dataLine.split(',', 1)[1]
#                 feature = json.loads(feature)
#                 string_tags = feature["string_tags"]
#                 feature = feature["vector"]["feature"]
#                 # print("feature:" + json.dumps(feature))
#                 data = querydata.replace("[0.9405091]", json.dumps(feature))
#                 # data = {
#                 #     "query": {
#                 #         "filter": [{
#                 #             "range": {
#                 #                 "int" : {
#                 #                     "gte" : 0,
#                 #                     "lte" : 9999
#                 #                 }
#                 #             }
#                 #         }],
#                 #         "sum" :[{
#                 #             "field": "vector",
#                 #             "feature": feature,
#                 #             "format":"normalization"
#                 #         }]
#                 #     }
#                 # }
#                 print("data:" + data)
#                 # response = requests.post(url, headers=headers, data=json.dumps(data))
#                 response = requests.post(url, headers=headers, data=data)
#                 print("searchByFeature---\n" + response.text)
#                 assert response.status_code == 200
#                 assert response.text.find("\"successful\":1")>=0

# def test_updateDoc():
#     logger.info("updateDoc")
#     headers = {"content-type": "application/json"}
#     with open(fileData, "r") as dataLine1:
#         for i, dataLine in zip(range(10),dataLine1):
#             idStr = dataLine.split(',', 1)[0].replace('{', '')
#             id = eval(idStr.split(':')[1])
#             data = {
#                 "doc":{
#                     "int": 32
#                 }
#             }
#             url = "http://" + ip_data + "/" + db_name + "/" + space_name_mmap + "/" + id + "/_update"
#             response = requests.post(url, headers=headers, data=data)
#             print("updateDoc:" + response.text)
#             assert response.status_code == 200

def test_deleteDocById():
    logger.info("test_deleteDoc")
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            data = "{"+dataLine.split(',', 1)[1]
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_mmap + "/" + id
            response = requests.delete(url)
            print("deleteDocById:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0") >= 0

def test_insertBulk():
    logger.info("insertBulk")
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_mmap+"/_bulk"
    headers = {"content-type": "application/json"}
    data = ''
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0]+"}"
            index = "{\"index\":"+idStr+"}"
            index = index + "\n"
            dataStr = "{"+dataLine.split(',', 1)[1]
            data = data + index + dataStr
        response = requests.post(url, headers=headers, data=data)
        print("insertBulk:" + response.text)
        assert response.status_code == 200

def test_deleteDocByIdBulk():
    logger.info("test_deleteDoc")
    headers = {"content-type": "application/json"}
    url = "http://" + ip_data + "/" + db_name + "/" + space_name_mmap + "/_bulk"
    with open(fileData, "r") as dataLine1:
        data = ''
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0]+"}"
            deleteid = "{\"delete\":"+idStr+"}\n"
            data = data + deleteid
        print("data:"+data)
        response = requests.post(url, headers=headers, data=data)
        print("deleteDocById:" + response.text)
        assert response.status_code == 200

def test_insterNoId1():
    logger.info("insertDataNoId")
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            data = "{"+dataLine.split(',', 1)[1]
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_mmap
            response = requests.post(url, headers=headers, data=data)
            print("insertNoID:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"successful\":1")>=0

def test_deleteDocByFeature():
    logger.info("test_deleteDoc")
    headers = {"content-type": "application/json"}
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_mmap+"/_delete_by_query"
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(1),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            # assert response.text.find("\"successful\":1")>=0

def test_deleteSpace():
    url = "http://" + ip_db + "/space/"+db_name+"/"+space_name_mmap
    response = requests.delete(url)
    print("deleteSpace:" + response.text)
    assert response.status_code == 200

# def test_deletetable():
#     url_table = "http://" + ip_db + "/space/" + db_name +"/_create"
#     url_delete = "http://" + ip_db + "/space/"+db_name+"/"+space_name_mmap
#     headers = {"content-type": "application/json"}
#     data_table = {
#         "name": "ts_space",
#         "dynamic_schema": "strict",
#         "partition_num": 3,
#         "replica_num": 1,
#         "engine": {"name": "gamma", "index_size": 10000, "max_size": 20000000},
#         "properties": {
#             "sku": {
#                 "type": "integer",
#                 "index": "false"
#             },
#             "img_url": {
#                 "type": "keyword",
#                 "index": "false"
#             },
#             "cid1": {
#                 "type": "integer",
#                 "index": "true"
#             },
#             "cid2": {
#                 "type": "integer",
#                 "index": "true"
#             },
#             "cid3": {
#                 "type": "integer",
#                 "index": "true"
#             },
#             "spu": {
#                 "type": "integer",
#                 "index": "false"
#             },
#             "brand_id": {
#                 "type": "integer",
#                 "index": "false"
#             },
#             "feature": {
#                 "type": "vector",
#                 "model_id": "img",
#                 "dimension": 512,
#                 # "retrieval_type": "GPU",
#                 "store_param": {"cache_size": 40960}
#             }
#         },
#         "models": [{
#             "model_id": "vgg16",
#             "fields": ["url"],
#             "out": "feature"
#         }]
#     }
#     # data_table = {
#     #     "name": space_name_mmap,
#     #     "dynamic_schema": "strict",
#     #     "partition_num": 2, #"partition_num": 2-6之间
#     #     "replica_num": 1,
#     #     "engine": {"name":"gamma", "index_size":8192, "max_size":10000},
#     #     "properties": {
#     #         "string": {
#     #             "type" : "keyword",
#     #             "index" : "true"
#     #         },
#     #         "int": {
#     #             "type": "integer",
#     #             "index" : "true"
#     #         },
#     #         "float": {
#     #             "type": "float",
#     #             "index" : "true"
#     #         },
#     #         "vector": {
#     #             "type": "vector",
#     #             "model_id": "img",
#     #             "dimension": 128,
#     #             "format":"normalization"
#     #         },
#     #         "string_tags": {
#     #             "type": "string",
#     #             "array": True,
#     #             "index" : "true"
#     #         },
#     #         "int_tags": {
#     #             "type": "integer",
#     #             "array": True,
#     #             "index" : "true"
#     #         },
#     #         "float_tags" : {
#     #             "type": "float",
#     #             "array": True,
#     #             "index" : "true"
#     #         }
#     #     },
#     #     "models": [{
#     #         "model_id": "vgg16",
#     #         "fields": ["string"],
#     #         "out": "feature"
#     #     }]
#     # }
#     f = open('result.txt', 'a+')
#     for i in range(200):
#         #create table
#         response = requests.put(url_table, headers=headers, data=json.dumps(data_table))
#         print("i---",i,":space_create", response.text)
#         #instert docs 1000
#         with open(fileData, "r") as dataLine1:
#             num = 0
#             for dataLine in dataLine1:
#                 idStr = dataLine.split(',', 1)[0].replace('{', '')
#                 id = eval(idStr.split(':')[1])
#                 data = "{"+dataLine.split(',', 1)[1]
#                 url = "http://" + ip_data + "/" + db_name + "/" + space_name_mmap + "/" + id
#                 # response = requests.post(url, headers=headers, data=data)
#
#                 # print("i---",i,"insertWithID:", response.text,file=f)
#                 # assert response.status_code == 200
#                 # num += 1
#                 # if(num == 1000):
#                 #     break;
#         #delete table
#         response = requests.delete(url_delete)
#         print("i---",i, ":delete space", response.text)
#         # time.sleep(5)
#     f.close()


logger.info("rocksdb")
def test_createspacerocksdb():
    url = "http://" + ip_db + "/space/" + db_name +"/_create"
    headers = {"content-type": "application/json"}
    data = {
        "name": space_name_rocksdb,
        "dynamic_schema": "strict",
        "partition_num": 1,  # "partition_num": 2-6之间
        "replica_num": 1,
        "engine": {
            "name":"gamma",
            "index_size": 400000,
            "max_size": 15000000,
            # "nprobe": 10,
            # "metric_type": "L2",
            # "ncentroids": 2048,
            # "nsubvector": 64
        },
        "properties": {
            "string": {
                "type" : "keyword",
                "index" : True
            },
            "int": {
                "type": "integer",
                "index" : True
            },
            "float": {
                "type": "float",
                "index" : True
            },
            "vector": {
                "type": "vector",
                "model_id": "img",
                "dimension": 128,
                "format":"normalization",
                # "retrieval_type": "GPU",
                "store_type": "RocksDB",
                "store_param":
                    {
                        "cache_size": 1024
                    }
            },
            "string_tags": {
                "type": "string",
                "array": True,
                "index" : True
            },
            "int_tags": {
                "type": "integer",
                "array": True,
                "index" : True
            },
            "float_tags" : {
                "type": "float",
                "array": True,
                "index" : True
            }
        },
        "models": [{
            "model_id": "vgg16",
            "fields": ["string"],
            "out": "feature"
        }]
    }
    # data = {
    #     "name": "ts_space",
    #     "dynamic_schema": "strict",
    #     "partition_num": 3,
    #     "replica_num": 3,
    #     "engine": {"name": "gamma", "index_size": 10000, "max_size": 20000000},
    #     "properties": {
    #         "sku": {
    #             "type": "integer",
    #             "index": "false"
    #         },
    #         "img_url": {
    #             "type": "keyword",
    #             "index": "false"
    #         },
    #         "cid1": {
    #             "type": "integer",
    #             "index": "true"
    #         },
    #         "cid2": {
    #             "type": "integer",
    #             "index": "true"
    #         },
    #         "cid3": {
    #             "type": "integer",
    #             "index": "true"
    #         },
    #         "spu": {
    #             "type": "integer",
    #             "index": "false"
    #         },
    #         "brand_id": {
    #             "type": "integer",
    #             "index": "false"
    #         },
    #         "feature": {
    #             "type": "vector",
    #             "model_id": "img",
    #             "dimension": 512,
    #             # "retrieval_type": "GPU",
    #             "store_param": {"cache_size":40960}
    #         }
    #     },
    #     "models": [{
    #         "model_id": "vgg16",
    #         "fields": ["url"],
    #         "out": "feature"
    #     }]
    # }
    print(url+"---"+json.dumps(data))
    response = requests.put(url, headers=headers, data=json.dumps(data))
    print("space_create---\n" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"msg\":\"success\"")>=0

def test_getspace_name_rocksdb():
    url = "http://" + ip_db + "/space/"+db_name+"/" + space_name_rocksdb
    response = requests.get(url)
    print("get_space---\n" + response.text)
    assert response.status_code == 200
    assert response.text.find("\"msg\":\"success\"")>=0


# def test_changemember():
#     url = "http://" + ip_db + "/partition/change_member"
#     headers = {"content-type": "application/json"}
#     data = {
#         "partition_id":7,
#         "node_id":1,
#         "method":0
#     }
#     response = requests.post(url, headers=headers, data=json.dumps(data))
#     print("change_member:" + response.text)
#     assert response.status_code == 200
#     assert response.text.find("\"msg\":\"success\"")>=0

logger.info("router(PS)")
def test_insertWithIdRocksdb():
    logger.info("insert")
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            flag = 0
            flag1 = idStr.split(':')[1].replace('\"','')
            id = str(int(flag1)+flag)
            data = "{"+dataLine.split(',', 1)[1]
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_rocksdb + "/" + id
            response = requests.post(url, headers=headers, data=data)
            print("insertWithID:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"status\":201")>=0

def test_searchByIdRocksdb():
    logger.info("test_searchById")
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            # id = eval(idStr.split(':')[1])
            flag = 0
            flag1 = idStr.split(':')[1].replace('\"','')
            id = str(int(flag1)+flag)
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_rocksdb + "/" + id
            response = requests.get(url)
            print("searchById:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"found\":true")>=0

def test_insterNoIdRocksdb():
    logger.info("insertDataNoId")
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            data = "{"+dataLine.split(',', 1)[1]
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_rocksdb
            response = requests.post(url, headers=headers, data=data)
            print("insertNoID:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"successful\":1")>=0

def test_searchByFeatureRocksdb():
    headers = {"content-type": "application/json"}
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_rocksdb+"/_search?size=100"
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0")>=0

def test_searchByFeatureandFilterRocksdb():
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_rocksdb+"/_search"
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            string_tags = feature["string_tags"]
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "filter": [{
                        "string_tags": string_tags
                    }],
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0") >= 0

def test_searchByFeatureandRangeRocksdb():
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_rocksdb+"/_search"
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(1),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            string_tags = feature["string_tags"]
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "filter": [{
                        "range": {
                            "int" : {
                                "gte" : 0,
                                "lte" : 0
                            }
                        }
                    }],
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            # print("data:" + json.dumps(data))
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0") >= 0

def test_searchByTermRocksdb():
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_rocksdb+"/_search"
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(1),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            string_tags = feature["string_tags"]
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "filter": [{
                        "term": {
                            "string": "0AW1mK_j19FyJvn5NR4Eb",
                            "operator": "or"
                        }
                    }],
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            # data = {
            #     "query": {
            #         "filter": [{
            #             "term": {
            #                 "string": "0AW1mK_j19FyJvn5NR4Eb"
            #             }
            #         }]
            #     }
            # }
            # print("data:" + json.dumps(data))
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0") >= 0

# def test_searchAll():
#     url = "http://" + ip_data + "/"+db_name+"/"+space_name_mmap+"/_search"
#     headers = {"content-type": "application/json"}
#     with open(fileData, "r") as dataLine1:
#         for querydata in query_list:
#             print("querydata:" + querydata)
#             for i, dataLine in zip(range(1),dataLine1):
#                 idStr = dataLine.split(',', 1)[0].replace('{', '')
#                 id = eval(idStr.split(':')[1])
#                 feature = "{"+dataLine.split(',', 1)[1]
#                 feature = json.loads(feature)
#                 string_tags = feature["string_tags"]
#                 feature = feature["vector"]["feature"]
#                 # print("feature:" + json.dumps(feature))
#                 data = querydata.replace("[0.9405091]", json.dumps(feature))
#                 # data = {
#                 #     "query": {
#                 #         "filter": [{
#                 #             "range": {
#                 #                 "int" : {
#                 #                     "gte" : 0,
#                 #                     "lte" : 9999
#                 #                 }
#                 #             }
#                 #         }],
#                 #         "sum" :[{
#                 #             "field": "vector",
#                 #             "feature": feature,
#                 #             "format":"normalization"
#                 #         }]
#                 #     }
#                 # }
#                 print("data:" + data)
#                 # response = requests.post(url, headers=headers, data=json.dumps(data))
#                 response = requests.post(url, headers=headers, data=data)
#                 print("searchByFeature---\n" + response.text)
#                 assert response.status_code == 200
#                 assert response.text.find("\"successful\":1")>=0

# def test_updateDocRocksdb():
#     logger.info("updateDoc")
#     headers = {"content-type": "application/json"}
#     with open(fileData, "r") as dataLine1:
#         for i, dataLine in zip(range(10),dataLine1):
#             idStr = dataLine.split(',', 1)[0].replace('{', '')
#             id = eval(idStr.split(':')[1])
#             data = {
#                 "doc":{
#                     "int": 32
#                 }
#             }
#             url = "http://" + ip_data + "/" + db_name + "/" + space_name_rocksdb + "/" + id + "/_update"
#             response = requests.post(url, headers=headers, data=data)
#             print("updateDoc:" + response.text)
#             assert response.status_code == 200

def test_deleteDocByIdRocksdb():
    logger.info("test_deleteDoc")
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            data = "{"+dataLine.split(',', 1)[1]
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_rocksdb + "/" + id
            response = requests.delete(url)
            print("deleteDocById:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"failed\":0") >= 0

def test_insertBulkRocksdb():
    logger.info("insertBulk")
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_rocksdb+"/_bulk"
    headers = {"content-type": "application/json"}
    data = ''
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0]+"}"
            index = "{\"index\":"+idStr+"}"
            index = index + "\n"
            dataStr = "{"+dataLine.split(',', 1)[1]
            data = data + index + dataStr
        response = requests.post(url, headers=headers, data=data)
        print("insertBulk:" + response.text)
        assert response.status_code == 200

def test_deleteDocByIdBulkRocksdb():
    logger.info("test_deleteDoc")
    headers = {"content-type": "application/json"}
    url = "http://" + ip_data + "/" + db_name + "/" + space_name_rocksdb + "/_bulk"
    with open(fileData, "r") as dataLine1:
        data = ''
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0]+"}"
            deleteid = "{\"delete\":"+idStr+"}\n"
            data = data + deleteid
        print("data:"+data)
        response = requests.post(url, headers=headers, data=data)
        print("deleteDocById:" + response.text)
        assert response.status_code == 200

def test_insterNoId1Rocksdb():
    logger.info("insertDataNoId")
    headers = {"content-type": "application/json"}
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(100),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            data = "{"+dataLine.split(',', 1)[1]
            url = "http://" + ip_data + "/" + db_name + "/" + space_name_rocksdb
            response = requests.post(url, headers=headers, data=data)
            print("insertNoID:" + response.text)
            assert response.status_code == 200
            assert response.text.find("\"successful\":1")>=0

def test_deleteDocByFeatureRocksdb():
    logger.info("test_deleteDoc")
    headers = {"content-type": "application/json"}
    url = "http://" + ip_data + "/"+db_name+"/"+space_name_rocksdb+"/_delete_by_query"
    with open(fileData, "r") as dataLine1:
        for i, dataLine in zip(range(1),dataLine1):
            idStr = dataLine.split(',', 1)[0].replace('{', '')
            id = eval(idStr.split(':')[1])
            feature = "{"+dataLine.split(',', 1)[1]
            feature = json.loads(feature)
            feature = feature["vector"]["feature"]
            data = {
                "query": {
                    "sum" :[{
                        "field": "vector",
                        "feature": feature,
                        "format":"normalization"
                    }]
                }
            }
            response = requests.post(url, headers=headers, data=json.dumps(data))
            print("searchByFeature---\n" + response.text)
            assert response.status_code == 200
            # assert response.text.find("\"successful\":1")>=0

def test_deleteSpaceRocksdb():
    url = "http://" + ip_db + "/space/"+db_name+"/"+space_name_rocksdb
    response = requests.delete(url)
    print("deleteSpace:" + response.text)
    assert response.status_code == 200

# def test_deletetableRocksdb():
#     url_table = "http://" + ip_db + "/space/" + db_name +"/_create"
#     url_delete = "http://" + ip_db + "/space/"+db_name+"/"+space_name_rocksdb
#     headers = {"content-type": "application/json"}
#     # data_table = {
#     #     "name": "ts_space",
#     #     "dynamic_schema": "strict",
#     #     "partition_num": 3,
#     #     "replica_num": 3,
#     #     "engine": {"name": "gamma", "index_size": 10000, "max_size": 20000000},
#     #     "properties": {
#     #         "sku": {
#     #             "type": "integer",
#     #             "index": "false"
#     #         },
#     #         "img_url": {
#     #             "type": "keyword",
#     #             "index": "false"
#     #         },
#     #         "cid1": {
#     #             "type": "integer",
#     #             "index": "true"
#     #         },
#     #         "cid2": {
#     #             "type": "integer",
#     #             "index": "true"
#     #         },
#     #         "cid3": {
#     #             "type": "integer",
#     #             "index": "true"
#     #         },
#     #         "spu": {
#     #             "type": "integer",
#     #             "index": "false"
#     #         },
#     #         "brand_id": {
#     #             "type": "integer",
#     #             "index": "false"
#     #         },
#     #         "feature": {
#     #             "type": "vector",
#     #             "model_id": "img",
#     #             "dimension": 512,
#     #             # "retrieval_type": "GPU",
#     #             "store_param": {"cache_size": 40960}
#     #         }
#     #     },
#     #     "models": [{
#     #         "model_id": "vgg16",
#     #         "fields": ["url"],
#     #         "out": "feature"
#     #     }]
#     # }
#     data_table = {
#         "name": space_name_rocksdb,
#         "dynamic_schema": "strict",
#         "partition_num": 3, #"partition_num": 2-6之间
#         "replica_num": 1,
#         "engine": {"name":"gamma", "index_size":8192, "max_size":100000},
#         "properties": {
#             "string": {
#                 "type" : "keyword",
#                 "index" : True
#             },
#             "int": {
#                 "type": "integer",
#                 "index" : True
#             },
#             "float": {
#                 "type": "float",
#                 "index" : True
#             },
#             "vector": {
#                 "type": "vector",
#                 "model_id": "img",
#                 "dimension": 128,
#                 "format":"normalization"
#             },
#             "string_tags": {
#                 "type": "string",
#                 "array": True,
#                 "index" : True
#             },
#             "int_tags": {
#                 "type": "integer",
#                 "array": True,
#                 "index" : True
#             },
#             "float_tags" : {
#                 "type": "float",
#                 "array": True,
#                 "index" : True
#             }
#         },
#         "models": [{
#             "model_id": "vgg16",
#             "fields": ["string"],
#             "out": "feature"
#         }]
#     }
#     f = open('result.txt', 'a+')
#     for i in range(200):
#         #create table
#         response = requests.put(url_table, headers=headers, data=json.dumps(data_table))
#         print("i---",i,":space_create", response.text)
#         #instert docs 1000
#         with open(fileData, "r") as dataLine1:
#             num = 0
#             for dataLine in dataLine1:
#                 idStr = dataLine.split(',', 1)[0].replace('{', '')
#                 id = eval(idStr.split(':')[1])
#                 data = "{"+dataLine.split(',', 1)[1]
#                 url = "http://" + ip_data + "/" + db_name + "/" + space_name_mmap + "/" + id
#                 # response = requests.post(url, headers=headers, data=data)
#
#                 # print("i---",i,"insertWithID:", response.text,file=f)
#                 # assert response.status_code == 200
#                 # num += 1
#                 # if(num == 1000):
#                 #     break;
#         #delete table
#         response = requests.delete(url_delete)
#         print("i---",i, ":delete space", response.text)
#         # time.sleep(5)
#     f.close()
#

def test_deleteDB():
    url = "http://" + ip_db + "/db/"+db_name
    response = requests.delete(url)
    print("deleteDB:" + response.text)
    assert response.status_code == 200
