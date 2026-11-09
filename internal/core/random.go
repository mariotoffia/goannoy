package core

type Kiss32 struct{ x, y, z, c uint32 }

func NewKiss32(seed uint32) *Kiss32 { return &Kiss32{seed, 362436000, 521288629, 7654321} }
func (r *Kiss32) Next() uint32 {
	r.x = 69069*r.x + 12345
	r.y ^= r.y << 13
	r.y ^= r.y >> 17
	r.y ^= r.y << 5
	v := 698769069*uint64(r.z) + uint64(r.c)
	r.c = uint32(v >> 32)
	r.z = uint32(v)
	return r.x + r.y + r.z
}

type Kiss64 struct{ x, y, z, c uint64 }

func NewKiss64(seed uint64) *Kiss64 {
	return &Kiss64{seed, 362436362436362436, 1066149217761810, 123456123456123456}
}
func (r *Kiss64) Next() uint64 {
	r.z = 6906969069*r.z + 1234567
	r.y ^= r.y << 13
	r.y ^= r.y >> 17
	r.y ^= r.y << 43
	v := (r.x << 58) + r.c
	r.c = r.x >> 6
	r.x += v
	if r.x < v {
		r.c++
	}
	return r.x + r.y + r.z
}
func (r *Kiss64) Index(n int) int { return int(r.Next() % uint64(n)) }
func (r *Kiss64) Flip() int       { return int(r.Next() & 1) }
