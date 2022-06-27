package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/hongping1224/lidario"
)

type pdis struct {
	index int
	dis   float64
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func main() {
	dir := flag.String("dir", "", "input Folder")
	size := flag.Float64("size", 0.2, "voxel Size(m)")
	out := flag.String("out", "./Output", "Output Folder")
	flag.Parse()
	fmt.Println(*size)
	inputRoot, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal("Fail to read Abs input Folder path")
		return
	}
	fileinfo, err := os.Stat(inputRoot)
	if os.IsNotExist(err) {
		log.Fatal("path does not exist.")
	}

	lasPaths := []string{inputRoot}
	if fileinfo.IsDir() {
		lasPaths = findFile(inputRoot, ".las")
	}
	for _, lasPath := range lasPaths {
		fmt.Println("start read las", lasPath)

		las, err := lidario.NewLasFile(lasPath, "r")
		if err != nil {
			log.Printf("%s read fail. Err : %v", las, err)
		}
		xmin := roundFloat(las.Header.MinX, 5)
		ymin := roundFloat(las.Header.MinY, 5)
		zmin := roundFloat(las.Header.MinZ, 5)
		xmax := roundFloat(las.Header.MaxX, 5)
		ymax := roundFloat(las.Header.MaxY, 5)
		zmax := roundFloat(las.Header.MaxZ, 5)
		xsize := int(math.Ceil((xmax-xmin) / *size)) + 1
		ysize := int(math.Ceil((ymax-ymin) / *size)) + 1
		zsize := int(math.Ceil((zmax-zmin) / *size)) + 1
		fmt.Println("finish read las")
		fmt.Println(xsize, ysize, zsize)
		remainPoints := make([][][]*pdis, xsize)

		fmt.Println("initial array")
		for i, _ := range remainPoints {
			remainPoints[i] = make([][]*pdis, ysize)
			for j, _ := range remainPoints[i] {
				remainPoints[i][j] = make([]*pdis, zsize)
				// for k, _ := range remainPoints[i][j] {
				// 	remainPoints[i][j][k] = -1
				// }
			}
		}
		fmt.Println("finish initial array")
		fmt.Println("start assign point")

		for i := 0; i < las.Header.NumberPoints; i++ {
			x, y, z, _ := las.GetXYZ(i)
			xc := int(math.Floor((x - xmin) / *size))
			yc := int(math.Floor((y - ymin) / *size))
			zc := int(math.Floor((z - zmin) / *size))
			centerx := ((float64(xc) + 0.5) * *size) + xmin
			centery := ((float64(yc) + 0.5) * *size) + ymin
			centerz := ((float64(zc) + 0.5) * *size) + zmin
			d1 := CalDistance(x, y, z, centerx, centery, centerz)
			if xc == -1 || yc == -1 || zc == -1 {
				continue
			}
			if remainPoints[xc][yc][zc] == nil {
				remainPoints[xc][yc][zc] = &pdis{index: i, dis: d1}
				continue
			}

			d2 := remainPoints[xc][yc][zc].dis
			if d1 < d2 {
				remainPoints[xc][yc][zc] = &pdis{index: i, dis: d1}
			}
		}
		fmt.Println("finish assign point")
		fmt.Println("start output point")

		outpath := filepath.Join(*out, filepath.Base(lasPath))
		outLas, err := lidario.InitializeUsingFile(outpath, las)
		for _, i := range remainPoints {
			for _, j := range i {
				for _, k := range j {
					if k == nil {
						continue
					}
					p, _ := las.LasPoint(k.index)
					outLas.AddLasPoint(p)
				}
			}
		}
		outLas.Close()
		fmt.Println("finish output point")
		las.Close()
	}

}

func CalDistance(x1, y1, z1, x2, y2, z2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2) + math.Pow(z1-z2, 2))
}

func findFile(root string, match string) (file []string) {

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if strings.HasSuffix(info.Name(), match) {
			file = append(file, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("Total shp file : ", len(file))
	return file
}
