#include "annoylib.h"
#include "kissrandom.h"
#include <fstream>
#include <iomanip>
#include <iostream>
#include <string>
using namespace Annoy;
constexpr int items = 64;
constexpr int trees = 3;
constexpr uint64_t seed = 42;
template<class Metric>
void floats(const std::string& dir, const char* metric, bool read) {
  AnnoyIndex<int32_t,float,Metric,Kiss64Random,AnnoyIndexSingleThreadedBuildPolicy> idx(3);
  std::string path=dir+"/"+metric+".ann";
  if (read) {
    if(!idx.load(path.c_str())) std::exit(2);
  } else {
    idx.set_seed(seed);
    for(int i=0;i<items;i++) {
      float v[3]={float((i*17)%97-48)/8, float((i*31)%89-44)/8, float((i*43)%83-41)/8};
      idx.add_item(i,v);
    }
    idx.build(trees,1);
    if(!idx.save(path.c_str())) std::exit(2);
  }
  float query[3]={1.25f,-2.5f,0.75f};
  std::vector<int32_t> ids; std::vector<float> dist;
  idx.get_nns_by_vector(query,10,10000,&ids,&dist);
  std::cout << "{\"metric\":\"" << metric << "\",\"dimension\":3,\"ids\":[";
  for(size_t i=0;i<ids.size();i++) std::cout<<(i?",":"")<<ids[i];
  std::cout << "],\"distances\":[";
  for(size_t i=0;i<dist.size();i++) std::cout<<(i?",":"")<<std::setprecision(9)<<dist[i];
  std::cout << "]}";
}
void hamming(const std::string& dir,bool read) {
  AnnoyIndex<int32_t,uint64_t,Hamming,Kiss64Random,AnnoyIndexSingleThreadedBuildPolicy> idx(3);
  auto path=dir+"/hamming.ann";
  if(read) { if(!idx.load(path.c_str())) std::exit(2); }
  else {
    idx.set_seed(seed);
    for(uint64_t i=0;i<items;i++) {
      uint64_t v[3]={i*0x9e3779b97f4a7c15ULL,(i+3)*0xd6e8feb86659fd93ULL,i&3};
      idx.add_item(i,v);
    }
    idx.build(trees,1); if(!idx.save(path.c_str())) std::exit(2);
  }
  uint64_t query[3]={0x123456789abcdef0ULL,0xfedcba9876543210ULL,1};
  std::vector<int32_t> ids; std::vector<uint64_t> dist;
  idx.get_nns_by_vector(query,10,10000,&ids,&dist);
  std::cout<<"{\"metric\":\"hamming\",\"dimension\":130,\"ids\":[";
  for(size_t i=0;i<ids.size();i++)std::cout<<(i?",":"")<<ids[i];
  std::cout<<"],\"distances\":[";
  for(size_t i=0;i<dist.size();i++)std::cout<<(i?",":"")<<dist[i];
  std::cout<<"]}";
}
int main(int argc,char**argv) {
  if(argc!=3) return 1;
  bool read=std::string(argv[1])=="read";
  std::cout<<"{\"upstream\":\"379f744667aba6b40ba3db8a07678df173a88f74\",\"seed\":42,\"trees\":3,\"items\":64,\"fixtures\":[";
  floats<Angular>(argv[2],"angular",read); std::cout<<",";
  floats<DotProduct>(argv[2],"dot",read); std::cout<<",";
  floats<Euclidean>(argv[2],"euclidean",read); std::cout<<",";
  floats<Manhattan>(argv[2],"manhattan",read); std::cout<<",";
  hamming(argv[2],read);
  std::cout<<"]}\n";
}
