package itemset;

import (
    "math"
    "math/rand"
    "time"
    "github.com/wenkesj/rphash/utils"
);

type KHHCountMinSketch struct {
    depth int;
    width int;
    table [depth][width]int64;
    hashA []int64;
    size int64;
    p *utils.PQueue64;
    k int;
    items map[int64]int64;
    countlist map[int64]int64;
    count int64;
    counts []int64;
    topcent []int64;
};

func NewKHHCountMinSketch(m int) *KHHCountMinSketch {
    k := int(float64(m) * math.Log(float64(m)));
    seed := int64(time.Now().UnixNano() / int64(time.Millisecond));
    countlist := make(map[int64]int64);
    items := make(map[int64]int64);
    var table [depth][width]int64;
    hashA := make([]int64, depth);
    random := rand.New(rand.NewSource(seed));
    for i := 0; i < depth; i++ {
        hashA[i] = random.Int63n(2147483647);
    }
    return &KHHCountMinSketch{
        k: k,
        countlist: countlist,
        hashA: hashA,
        items: items,
        table: table,
        width: width,
        depth: depth,
        size: 0,
        p: utils.NewPQueue64(),
        topcent: nil,
    };
};

func (this *KHHCountMinSketch) Hash(item int64, i int) int {
    const PRIME_MODULUS = (int64(1) << 31) - 1;
    hash := this.hashA[i] * item;
    hash += hash >> 32;
    hash &= PRIME_MODULUS;
    return int(hash) % width;
};

func (this *KHHCountMinSketch) Add(e int64) {
    count := this.AddLong(utils.HashCode(e), 1);
    if _, ok := this.items[utils.HashCode(e)]; !ok {
        this.countlist[utils.HashCode(e)] = count;
        this.p.Add(e);
        this.items[utils.HashCode(e)] = e;
    } else {
        this.p.Remove(e);
        this.items[utils.HashCode(e)] = e;
        this.countlist[utils.HashCode(e)] = count;
        this.p.Add(e);
    }

    if this.p.Size() > this.k {
        removed := this.p.Poll();
        delete(this.items, removed);
    }
};

func (this *KHHCountMinSketch) AddLong(item, count int64) int64 {
    this.table[0][this.Hash(item, 0)] += count;
    min := int64(this.table[0][this.Hash(item, 0)]);
    for i := 1; i < this.depth; i++ {
        this.table[i][this.Hash(item, i)] += count;
        if this.table[i][this.Hash(item, i)] < min {
            min = int64(this.table[i][this.Hash(item, i)]);
        }
    }
    this.size += count;
    return min;
};

func (this *KHHCountMinSketch) GetCount() int64 {
    return this.count;
};

func (this *KHHCountMinSketch) GetCounts() []int64 {
    if this.counts != nil {
        return this.counts;
    }
    this.GetTop();
    return this.counts;
};

func (this *KHHCountMinSketch) GetTop() []int64 {
    if this.topcent != nil {
        return this.topcent;
    }
    this.topcent = []int64{};
    this.counts = []int64{};
    for !this.p.IsEmpty() {
        tmp := this.p.Poll();
        this.topcent = append(this.topcent, tmp);
        this.counts = append(this.counts, this.countlist[utils.HashCode(tmp)]);
    }
    return this.topcent;
};
