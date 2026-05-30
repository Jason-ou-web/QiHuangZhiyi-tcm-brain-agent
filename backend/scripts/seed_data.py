"""Seed sample agricultural knowledge into Qdrant for MVP testing."""
import uuid
import requests
from qdrant_client import QdrantClient
from qdrant_client.models import PointStruct

docs = [
    {
        "content": "小麦锈病是小麦生产中的重要病害，主要分为条锈病、叶锈病和秆锈病三种。条锈病主要危害叶片，也可危害叶鞘、茎秆和穗部。防治方法包括选用抗病品种、合理施肥、药剂防治等。药剂防治可在发病初期喷施三唑酮、戊唑醇等杀菌剂。",
        "book_title": "小麦病虫害防治手册",
        "chapter": "第三章 小麦真菌性病害",
        "page": 45,
    },
    {
        "content": "水稻稻瘟病是水稻三大病害之一，整个生育期均可发生。根据危害部位不同，分为苗瘟、叶瘟、节瘟、穗颈瘟和谷粒瘟。穗颈瘟对产量影响最大。防治策略应以种植抗病品种为基础，结合栽培管理和药剂防治。常用药剂有稻瘟灵、三环唑、嘧菌酯等。",
        "book_title": "水稻病虫害图鉴",
        "chapter": "第二章 稻瘟病",
        "page": 32,
    },
    {
        "content": "玉米螟是玉米的主要害虫，以幼虫钻蛀危害。幼虫孵化后先取食叶片，后钻入茎秆、雌穗和雄穗危害。防治适期为大喇叭口期，可采用辛硫磷颗粒剂灌心或释放赤眼蜂进行生物防治。",
        "book_title": "玉米高产栽培技术",
        "chapter": "第五章 玉米虫害防治",
        "page": 78,
    },
    {
        "content": "土壤有机质是土壤肥力的核心指标。提高土壤有机质含量的主要措施包括：增施有机肥（堆肥、厩肥）、秸秆还田、种植绿肥、实行合理轮作等。一般农田土壤有机质含量以2%-3%为宜。",
        "book_title": "土壤肥料学",
        "chapter": "第四章 土壤有机质",
        "page": 62,
    },
    {
        "content": "番茄晚疫病是一种毁灭性病害，由致病疫霉菌引起。主要危害叶片、茎秆和果实。叶片受害初期呈水渍状斑点，湿度大时叶背产生白色霉层。防治措施：选用抗病品种、合理轮作、加强通风、药剂防治（代森锰锌、霜脲·锰锌等）。",
        "book_title": "蔬菜病虫害防治大全",
        "chapter": "第七章 茄科蔬菜病害",
        "page": 124,
    },
    {
        "content": "农用拖拉机按用途可分为通用型、中耕型和园艺型。按行走方式可分为轮式、履带式和半履带式。选购拖拉机需考虑功率、牵引力、燃油经济性及当地地块条件。日常保养包括定期更换机油、检查滤清器、保持轮胎气压等。",
        "book_title": "现代农业机械使用手册",
        "chapter": "第一章 拖拉机",
        "page": 8,
    },
    {
        "content": "《齐民要术》为北魏贾思勰所著，是中国现存最早最完整的农书。全书十卷九十二篇，系统总结了六世纪前黄河中下游地区农牧业生产经验。书中记载了谷物、蔬菜、果树、林木的栽培方法，以及家畜、家禽、鱼类的养殖技术，体现了顺天时、量地利的农业思想。",
        "book_title": "齐民要术",
        "chapter": "序言",
        "page": 1,
    },
    {
        "content": "棉花枯萎病是棉花生产中的主要病害之一，由尖孢镰刀菌引起，为典型的土传维管束病害。病株表现为叶片黄化、萎蔫，茎秆剖面维管束变褐色。防治上以选用抗病品种为主，配合轮作倒茬（与非寄主作物轮作5年以上）和种子处理。",
        "book_title": "经济作物病虫害防治",
        "chapter": "第三章 棉花病害",
        "page": 89,
    },
    {
        "content": "水稻田间水分管理要掌握浅水插秧、深水返青、薄水分蘖、够苗晒田、寸水孕穗、干湿壮籽的原则。晒田可以控制无效分蘖，促进根系下扎，增强抗倒伏能力。一般在分蘖末期至拔节前进行晒田，晒至田面出现鸡爪裂为宜。",
        "book_title": "水稻栽培学",
        "chapter": "第六章 水分管理",
        "page": 142,
    },
    {
        "content": "农作物轮作的原则：深根作物与浅根作物轮作，禾本科与豆科轮作，耗地作物与养地作物轮作。例如：玉米-大豆轮作可以利用大豆的固氮作用提高土壤氮素含量；水稻-油菜轮作可以改善土壤物理性状。合理的轮作制度可以减轻病虫害、平衡土壤养分、提高作物产量。",
        "book_title": "耕作学原理",
        "chapter": "第三章 轮作制度",
        "page": 55,
    },
]

print(f"Embedding {len(docs)} documents ...")
resp = requests.post("http://127.0.0.1:8081/embed", json={"texts": [d["content"] for d in docs]})
resp.raise_for_status()
embeddings = resp.json()["embeddings"]
print(f"Got {len(embeddings)} embeddings, dim={len(embeddings[0])}")

client = QdrantClient(host="localhost", port=6334, grpc_port=6334, prefer_grpc=True)

points = []
for i, (doc, vec) in enumerate(zip(docs, embeddings)):
    points.append(PointStruct(
        id=str(uuid.uuid4()),
        vector=vec,
        payload={
            "content": doc["content"],
            "book_title": doc["book_title"],
            "chapter": doc["chapter"],
            "page": doc["page"],
        },
    ))

client.upsert(collection_name="agri_knowledge", points=points)
print(f"Indexed {len(points)} documents successfully")

# verify
count = client.count(collection_name="agri_knowledge")
print(f"Total points in collection: {count.count}")
