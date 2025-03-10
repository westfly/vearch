/**
 * Copyright 2019 The Gamma Authors.
 *
 * This source code is licensed under the Apache License, Version 2.0 license
 * found in the LICENSE file in the root directory of this source tree.
 */

#ifndef GAMMA_API_H_
#define GAMMA_API_H_

#ifdef __cplusplus
extern "C" {
#endif

/** init an engine pointer
 *
 * @param config  engine config pointer
 * @return engine pointer
 */
void *Init(const char *config_str, int len);

/** destroy an engine point
 *
 * @param engine  a pointer to search engine
 * @return 0 successed, 1 failed
 */
int Close(void *engine);

/** create a table
 *
 * @param engine  search engine pointer
 * @param table   table info
 * @return 0 successed, 1 failed
 */
int CreateTable(void *engine, const char *table_str, int len);

/** add a doc to table, if doc existed, update it
 *
 * @param engine  search engine pointer
 * @param doc     doc pointer to add
 * @return 0 successed, 1 failed
 */
int AddOrUpdateDoc(void *engine, const char *doc_str, int len);

/** For batch operator
 * set add or update docs number
 *
 * @param engine  search engine pointer
 * @param i       docs number
 * @return 0 successed, 1 failed
 */
int AddOrUpdateDocsNum(void *engine, int i);

/** For batch operator
 * Prepare the docs to be batch updated
 *
 * @param engine    search engine pointer
 * @param doc_str   doc's serialized string
 * @param id        doc's idx
 * @return 0 successed, 1 failed
 */
int PrepareDocs(void *engine, char *doc_str, int id);

/** For batch operator
 * Batch add or update prepared docs
 *
 * @param engine     search engine pointer
 * @param len        doc's number
 * @param result_str result's serialized string
 * @param result_len result number
 * @return 0 successed, 1 failed
 */
int AddOrUpdateDocsFinish(void *engine, int len, char **result_str,
                          int *result_len);

/** For batch operator
 * Batch add or update docs
 * 
 * @param engine     search engine pointer
 * @param doc_str    docs' serialized string
 * @param len        docs number
 * @param result_str result's serialized string
 * @param result_str result's len
 * @return 0 successed, 1 failed
 */
int AddOrUpdateDocs(void *engine, char **doc_str, int len, char **result_str,
                    int *result_len);

/** update a doc, if _id not exist, equal to function @AddDoc
 *  the interface is not supported temporarily.
 * @param engine  search engine pointer
 * @param doc     doc pointer to update
 * @return 0 successed, 1 failed
 */
int UpdateDoc(void *engine, const char *doc_str, int len);

/** delete a doc from table
 *
 * @param engine  search engine pointer
 * @param doc     doc pointer to delete
 * @return 0 successed, 1 failed
 */
int DeleteDoc(void *engine, const char *docid, int docid_len);

/** get the engine status
 *
 * @param engine  search engine pointer
 */
void GetEngineStatus(void *engine, char **status, int *len);

void GetMemoryInfo(void *engine, char **memory_info, int *len);

/** get a doc by id
 *
 * @param engine
 * @param docid  doc id
 * @return  a doc
 */
int GetDocByID(void *engine, const char *docid, int docid_len, char **doc_str,
               int *len);

/** get a doc by docid
 *
 * @param engine
 * @param docid  doc id
 * @return  a doc
 */
int GetDocByDocID(void *engine, int docid, char **doc_str, int *len);

/** @param engine  search engine pointer
 * @return 0 successed, 1 failed
 */
int BuildIndex(void *engine);

/** @param engine  search engine pointer
 * @return 0 successed, 1 failed
 */
int RebuildIndex(void *engine, int drop_before_rebuild, int limit_cpu, int describe);

/** dump datas into disk accord to Config
 *
 * @param engine
 * @return 0 successed, 1 failed
 */
int Dump(void *engine);

/** load datas from disk accord to Config
 *
 * @param engine
 * @return 0 successed, 1 failed
 */
int Load(void *engine);

/** query vectors to index with serialized result
 *
 * @param engine    search engine pointer
 * @param request   search request pointer
 * @return response
 */
int Search(void *engine, const char *request_str, int req_len,
           char **response_str, int *res_len);

/** alter all cache size by query
 *
 * @param engine  search engine pointer
 * @param cache_str caches' serialized string
 * @return 0 successed, 1 failed
 */
int SetConfig(void *engine, const char *config_str, int len);

/** get all cache size by query
 *
 * @param engine  search engine pointer
 * @param cache_str caches' serialized string
 * @return 0 successed, 1 failed
 */
int GetConfig(void *engine, char **config_str, int *len);

#ifdef __cplusplus
}
#endif

#endif /* GAMMA_API_H */
