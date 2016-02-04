package simple;

import (
    "math"
    "github.com/wenkesj/rphash/types"
    "github.com/wenkesj/rphash/defaults"
);

type Simple struct {
    centroids [][]float64;
    variance float64;
    rphashObject types.RPHashObject;
};

func NewSimple(_rphashObject types.RPHashObject) *Simple {
    return &Simple{
        variance: 0,
        centroids: nil,
        rphashObject: _rphashObject,
    };
};

// Map is doing the count.
func (this *Simple) Map() *Simple {
    vecs := this.rphashObject.GetVectorIterator();
    var hashResult int64;
    vec := vecs.Next();
    hash := defaults.NewHash(this.rphashObject.GetHashModulus());
    decoder := this.rphashObject.GetDecoderType();
    // Here we are using the same input and target dimension... Why?
    projector := defaults.NewProjector(this.rphashObject.GetDimensions(), decoder.GetDimensionality(), this.rphashObject.GetRandomSeed());
    LSH := defaults.NewLSH(hash, decoder, projector);
    k := int(float64(this.rphashObject.GetK()) * math.Log(float64(this.rphashObject.GetK())));
    CountMinSketch := defaults.NewCountMinSketch(k);
    for vecs.HasNext() {
        // Project the Vector to lower dimension.
        // Decode the new vector for meaningful integers
        // Hash the new vector into a 64 bit int.
        hashResult = LSH.LSHHashSimple(vec);
        // Add it to the count min sketch to update frequencies.
        CountMinSketch.Add(hashResult);
        vec = vecs.Next();
    }
    this.rphashObject.SetPreviousTopID(CountMinSketch.GetTop());
    vecs.Reset();
    return this;
};

// Reduce is finding out where the centroids are in respect to the real data.
func (this *Simple) Reduce() *Simple {
    vecs := this.rphashObject.GetVectorIterator();
    if !vecs.HasNext() {
        return this;
    }

    hash := defaults.NewHash(this.rphashObject.GetHashModulus());
    decoder := this.rphashObject.GetDecoderType();
    projector := defaults.NewProjector(this.rphashObject.GetDimensions(), decoder.GetDimensionality(), this.rphashObject.GetRandomSeed());
    LSH := defaults.NewLSH(hash, decoder, projector);

    var centroids []types.Centroid;
    vec := vecs.Next();
    for i := 0; i < this.rphashObject.GetK(); i++ {
        // Get the top centroids.
        previousTop := this.rphashObject.GetPreviousTopID();
        centroid := defaults.NewCentroidSimple(this.rphashObject.GetDimensions(), previousTop[i]);
        centroids = append(centroids, centroid);
    }
    // Iterate over the dataset and check CountMinSketch.
    for vecs.HasNext() {
        var hashResult = LSH.LSHHashSimple(vec);
        // For each vector, check if it is a centroid.
        for _, cent := range centroids {
            // Get an idea where the LSH is in respect to the vector.
            if cent.GetIDs().Contains(hashResult) {
                cent.UpdateVector(vec);
                break;
            }
        }
        vec = vecs.Next();
    }
    for _, cent := range centroids {
        this.rphashObject.AddCentroid(cent.Centroid());
    }
    vecs.Reset();
    return this;
};

func (this *Simple) GetCentroids() [][]float64 {
    if this.centroids == nil {
        this.Run();
    }
    // Perform the KMeans on the centroids.
    return defaults.NewKMeansSimple(this.rphashObject.GetK(), this.centroids).GetCentroids();
};

func (this *Simple) Run() {
    this.Map().Reduce();
    this.centroids = this.rphashObject.GetCentroids();
}

func (this *Simple) GetRPHash() types.RPHashObject {
    return this.rphashObject;
};
