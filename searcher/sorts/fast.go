package sorts

import (
	"github.com/RoaringBitmap/roaring"
	"gofound/searcher/model"
	"gofound/searcher/utils"
	"log"
	"sort"
	"sync"
)

const (
	DESC = "desc"
	ASC  = "asc"
)

type ScoreSlice []model.SliceItem

func (x ScoreSlice) Len() int {
	return len(x)
}
func (x ScoreSlice) Less(i, j int) bool {
	return x[i].Score < x[j].Score
}
func (x ScoreSlice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

type Uint32Slice []uint32

func (x Uint32Slice) Len() int           { return len(x) }
func (x Uint32Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Uint32Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type FastSort struct {
	sync.Mutex

	IsDebug bool

	Keys []uint32

	bitmap *roaring.Bitmap

	Call func(keys []uint32, id uint32) float32
}

func (f *FastSort) Add(ids []uint32) {
	f.Lock()
	defer f.Unlock()
	if f.bitmap == nil {
		f.bitmap = roaring.BitmapOf(ids...)
	} else {
		f.bitmap.AddMany(ids)
	}
}

// Count 获取数量
func (f *FastSort) Count() int {
	return int(f.bitmap.GetCardinality())
}

func (f *FastSort) GetAll(order string) []model.SliceItem {

	//声明大小，避免重复合并数组
	var ids = f.bitmap.ToArray()

	var result = make([]model.SliceItem, len(ids))

	//降序排序
	_tt := utils.ExecTime(func() {

		if order == DESC {
			sort.Sort(sort.Reverse(Uint32Slice(ids)))
		}
	})
	if f.IsDebug {
		log.Println("排序 time:", _tt)
	}

	//计算相关度
	_tt = utils.ExecTime(func() {
		wg := sync.WaitGroup{}
		wg.Add(len(ids))

		for i, id := range ids {
			go func() {
				result[i] = model.SliceItem{
					Id:    id,
					Score: f.Call(f.Keys, id),
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
	if f.IsDebug {
		log.Println("计算相关度 time:", _tt)
	}
	//对分数进行排序
	sort.Sort(sort.Reverse(ScoreSlice(result)))

	return result
}
