/**
 * Copyright 2019 The Gamma Authors.
 *
 * This source code is licensed under the Apache License, Version 2.0 license
 * found in the LICENSE file in the root directory of this source tree.
 */

#include "test.h"

/**
 * To run this demo, please download the ANN_SIFT10K dataset from
 *
 *   ftp://ftp.irisa.fr/local/texmex/corpus/siftsmall.tar.gz
 *
 * and unzip it.
 **/

namespace test {

class GammaTest : public ::testing::Test {
 public:
  static int Init(int argc, char *argv[]) {
    GammaTest::my_argc = argc;
    GammaTest::my_argv = argv;
    return 0;
  }

 protected:
  GammaTest() {}

  ~GammaTest() override {}

  void SetUp() override {}

  void TearDown() override {}

  void *engine;

  static int my_argc;
  static char **my_argv;
};

int GammaTest::my_argc = 0;
char **GammaTest::my_argv = nullptr;

TEST_F(GammaTest, IVFPQ) {
  struct Options opt;
  opt.profile_file = my_argv[1];
  opt.feature_file = my_argv[2];
  opt.retrieval_type = "IVFPQ";
  ASSERT_EQ(TestIndexes(opt), 0);
  opt.b_load = true;
  ASSERT_EQ(TestIndexes(opt), 0);
}

int CreateFaissTable(struct Options &opt) {
  tig_gamma::TableInfo table;
  table.SetName(opt.vector_name);
  table.SetRetrievalType(opt.retrieval_type);
  if (opt.retrieval_type == "IVFPQ" || opt.retrieval_type == "IVFPQ_RELAYOUT") {
    table.SetRetrievalParam(kIVFPQParam);
  } else if (opt.retrieval_type == "IVFFLAT") {
    table.SetRetrievalParam(kIVFPQParam);
  } else if (opt.retrieval_type == "FLAT") {
    table.SetRetrievalParam(kFLATParam);
  } else if (opt.retrieval_type == "HNSW") {
    table.SetRetrievalParam(kHNSWParam);
  }

  table.SetIndexingSize(opt.indexing_size);

  struct tig_gamma::FieldInfo field_info;
  field_info.name = "_id";

  field_info.is_index = false;
  field_info.data_type = tig_gamma::DataType::STRING;
  table.AddField(field_info);

  struct tig_gamma::VectorInfo vector_info;
  vector_info.name = "faiss";
  vector_info.data_type = tig_gamma::DataType::FLOAT;
  vector_info.is_index = true;
  vector_info.dimension = opt.d;
  vector_info.model_id = opt.model_id;
  vector_info.store_type = opt.store_type;
  vector_info.store_param = "{\"cache_size\": 16, \"compress\": {\"rate\":16}}";
  // vector_info.store_param = "{\"cache_size\": 16}";

  table.AddVectorInfo(vector_info);

  char *table_str = nullptr;
  int len = 0;
  table.Serialize(&table_str, &len);

  int ret = CreateTable(opt.engine, table_str, len);

  free(table_str);

  return ret;
}

int DumpFaissIndex(struct Options &opt) {
  // dimension of the vectors to index
  int d = 512;

  // size of the database we plan to index
  size_t nb = 10000;

  // make a set of nt training vectors in the unit cube
  // (could be the database)
  size_t nt = 15000;

  // make the index object and train it
  faiss::IndexFlatL2 coarse_quantizer(d);

  // a reasonable number of cetroids to index nb vectors
  int ncentroids = 256;

  faiss::IndexIVFPQ index(&coarse_quantizer, d, ncentroids, 64, 8);

  // index that gives the ground-truth
  faiss::IndexFlatL2 index_gt(d);

  std::mt19937 rng;
  std::uniform_real_distribution<> distrib;

  {  // training

    std::vector<float> trainvecs(nt * d);
    for (size_t i = 0; i < nt * d; i++) {
      trainvecs[i] = distrib(rng);
    }
    index.verbose = true;
    index.train(nt, opt.feature);
  }

  {  // populating the database

    std::vector<float> database(nb * d);
    for (size_t i = 0; i < nb * d; i++) {
      database[i] = distrib(rng);
    }

    index.add(nb, opt.feature);
    index_gt.add(nb, opt.feature);
  }

  int nq = 200;
  int n_ok;

  {  // searching the database

    std::vector<float> queries(nq * d);
    for (int i = 0; i < nq * d; i++) {
      queries[i] = distrib(rng);
    }

    std::vector<faiss::Index::idx_t> gt_nns(nq);
    std::vector<float> gt_dis(nq);

    index_gt.search(nq, opt.feature, 1, gt_dis.data(), gt_nns.data());

    index.nprobe = 5;
    int k = 5;
    std::vector<faiss::Index::idx_t> nns(k * nq);
    std::vector<float> dis(k * nq);

    index.search(nq, opt.feature, k, dis.data(), nns.data());

    n_ok = 0;
    for (int q = 0; q < nq; q++) {
      for (int i = 0; i < k; i++)
        if (nns[q * k + i] == gt_nns[q]) n_ok++;
    }
    // EXPECT_GT(n_ok, nq * 0.4);
  }
  utils::make_dir("files/faiss.000");
  faiss::write_index(&index, "files/faiss.000/ivfpq.index");

  return 0;
}


int SearchFaiss(struct Options &opt, size_t num) {
  size_t idx = 0;
  double time = 0;
  int failed_count = 0;
  int req_num = 1;
  string error;
  while (idx < num) {
    double start = utils::getmillisecs();
    struct tig_gamma::VectorQuery vector_query;
    vector_query.name = "faiss";

    int len = opt.d * sizeof(float) * req_num;
    char *value = reinterpret_cast<char *>(opt.feature + (uint64_t)idx * opt.d);

    vector_query.value = std::string(value, len);

    vector_query.min_score = -10000;
    vector_query.max_score = 10000;
    vector_query.boost = 0.1;
    vector_query.has_boost = 0;

    tig_gamma::Request request;
    request.SetTopN(10);
    request.AddVectorQuery(vector_query);
    request.SetReqNum(req_num);
    request.SetBruteForceSearch(0);
    request.SetHasRank(true);
    std::string retrieval_params =
        "{\"metric_type\" : \"L2\", \"recall_num\" : "
        "10, \"nprobe\" : 10, \"ivf_flat\" : 0}";
    request.SetRetrievalParams(retrieval_params);
    // request.SetOnlineLogLevel("");
    request.SetMultiVectorRank(0);
    request.SetL2Sqrt(false);

    char *request_str, *response_str;
    int request_len, response_len;

    request.Serialize(&request_str, &request_len);
    int ret = Search(opt.engine, request_str, request_len, &response_str,
                     &response_len);

    if (ret != 0) {
      LOG(ERROR) << "Search error [" << ret << "]";
    }
    free(request_str);

    tig_gamma::Response response;
    response.Deserialize(response_str, response_len);

    free(response_str);

    if (opt.print_doc) {
      std::vector<struct tig_gamma::SearchResult> &results = response.Results();
      for (size_t i = 0; i < results.size(); ++i) {
        int ii = idx + i;
        string msg = std::to_string(ii) + ", ";
        struct tig_gamma::SearchResult &result = results[i];

        std::vector<struct tig_gamma::ResultItem> &result_items =
            result.result_items;
        if (result_items.size() <= 0) {
          LOG(ERROR) << "search no result, id=" << ii;
          continue;
        }
        msg += string("total [") + std::to_string(result.total) + "], ";
        msg += string("result_num [") + std::to_string(result_items.size()) +
               "], ";
        for (size_t j = 0; j < result_items.size(); ++j) {
          struct tig_gamma::ResultItem &result_item = result_items[j];
          printDoc(result_item, msg, opt);
          msg += "\n";
        }
        LOG(INFO) << msg;
        if (abs(result_items[0].score - 1.0) < 0.001) {
          if (ii % 100 == 0) {
            LOG(INFO) << msg;
          }
        } else {
          if (!bitmap::test(opt.docids_bitmap_, ii)) {
            LOG(ERROR) << msg;
            error += std::to_string(ii) + ",";
            bitmap::set(opt.docids_bitmap_, ii);
            failed_count++;
          }
        }
      }
    }
    double elap = utils::getmillisecs() - start;
    time += elap;
    if (idx % 10000 == 0) {
      LOG(INFO) << "search time [" << time / 10000 << "]ms";
      time = 0;
    }
    idx += req_num;
    if (idx >= opt.doc_id) {
      idx = 0;
      break;
    }
  }
  LOG(ERROR) << error;
  return failed_count;
}

TEST_F(GammaTest, LOAD_FAISS_INDEX) {
  struct Options opt;
  opt.profile_file = my_argv[1];
  opt.feature_file = my_argv[2];
  opt.retrieval_type = "IVFPQ";
  opt.store_type = "MemoryOnly";
  opt.add_doc_num = 20000;
  opt.indexing_size = 5000;
  opt.b_load = true;
  int ret = 0;
  InitEngine(opt);
  DumpFaissIndex(opt);
  CreateFaissTable(opt);
  int bitmap_bytes_size = 0;
  ret = bitmap::create(opt.docids_bitmap_, bitmap_bytes_size, opt.max_doc_size);
  if (ret != 0) {
    LOG(ERROR) << "Create bitmap failed!";
  }
  ret = Load(opt.engine);
  SearchFaiss(opt, 100);

  CloseEngine(opt);
}

}  // namespace test

int main(int argc, char **argv) {
  setvbuf(stdout, (char *)NULL, _IONBF, 0);
  ::testing::InitGoogleTest(&argc, argv);
  if (argc != 3) {
    std::cout << "Usage: [Program] [profile_file] [vectors_file]\n";
    return 1;
  }
  ::testing::GTEST_FLAG(output) = "xml";
  test::GammaTest::Init(argc, argv);
  return RUN_ALL_TESTS();
}
