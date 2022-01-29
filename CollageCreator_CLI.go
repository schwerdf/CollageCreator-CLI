// This package provides a command-line interface to the CollageCreator library.
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/schwerdf/CollageCreator"
)

var AllCollageCreatorComponents map[string]CollageCreator.CollageCreatorComponent = map[string]CollageCreator.CollageCreatorComponent{}

func init() {
	AllCollageCreatorComponents["ProgressMonitor"] = CollageCreator.ProgressMonitor_Init()
	AllCollageCreatorComponents["InputImageReader_Raster"] = CollageCreator.InputImageReader_Raster_Init()
	AllCollageCreatorComponents["DimensionInitializer_Uniform"] = CollageCreator.DimensionInitializer_Uniform_Init()
	AllCollageCreatorComponents["PositionCalculator_Random"] = CollageCreator.PositionCalculator_Random_Init()
	AllCollageCreatorComponents["PositionCalculator_TileInOrder"] = CollageCreator.PositionCalculator_TileInOrder_Init()
	AllCollageCreatorComponents["CollageRenderer_Raster"] = CollageCreator.CollageRenderer_Raster_Init()
	AllCollageCreatorComponents["CollageRenderer_SVG"] = CollageCreator.CollageRenderer_SVG_Init()
	AllCollageCreatorComponents["CollageRenderer_ImageMagickScript"] = CollageCreator.CollageRenderer_ImageMagickScript_Init()
}

func main() {
	var params = CollageCreator.Parameters_init()
	outFile := flag.String("o", "Collage", "Output to this image")
	padding := flag.String("padding", "0x0", "Require this much padding around images (relative to the canvas size) and the sides of the canvas")
	aspectRatio := flag.String("aspect", "0x0", "Aspect ratio of output canvas (0.0 to set automatically)")
	minCanvasSize := flag.String("minsize", "0x0", "Minimum size of output canvas (determined automatically by default)")
	maxCanvasSize := flag.String("maxsize", "0x0", "Maximum size of output canvas (determined automatically by default)")
	inputImageReader := flag.String("lc", "Raster", "Use this method to read input images")
	dimensionInitializer := flag.String("di", "Uniform", "Use this method to scale images before positioning")
	positionCalculator := flag.String("pc", "Random", "Use this method to position images")
	outputFileType := flag.String("t", "", " (default is to determine by output file extension)")
	valid := true
	for _, css := range AllCollageCreatorComponents {
		valid = valid && css.RegisterCustomParameters(&params)
	}
	if !valid {
		log.Fatal("Error registering custom parameters")
	}
	flag.Parse()

	inputImages := os.Args[len(os.Args)-flag.NArg() : len(os.Args)]

	if len(inputImages) < 1 {
		log.Fatal("Must have at least one input image")
	}

	params.SetProgressMonitor(AllCollageCreatorComponents["ProgressMonitor"].(CollageCreator.ProgressMonitor))

	if iirSwitch, ok := AllCollageCreatorComponents["InputImageReader_"+*inputImageReader]; ok {
		params.SetInputImageReader(iirSwitch.(CollageCreator.InputImageReader))
	} else {
		log.Fatal("Invalid InputImageReader:", *inputImageReader)
	}
	if diSwitch, ok := AllCollageCreatorComponents["DimensionInitializer_"+*dimensionInitializer]; ok {
		params.SetDimensionInitializer(diSwitch.(CollageCreator.DimensionInitializer))
	} else {
		log.Fatal("Invalid DimensionInitializer:", *dimensionInitializer)
	}
	if pcSwitch, ok := AllCollageCreatorComponents["PositionCalculator_"+*positionCalculator]; ok {
		params.SetPositionCalculator(pcSwitch.(CollageCreator.PositionCalculator))
	} else {
		log.Fatal("Invalid PositionCalculator:", *positionCalculator)
	}
	if *outputFileType == "" {
		ext := strings.ToLower(filepath.Ext(*outFile))
		if ext == "" {
			ext = strings.ToLower(filepath.Ext(inputImages[0]))
		}
		if ext == "" {
			*outputFileType = "jpg"
			*outFile = *outFile + ".jpg"
		} else {
			*outputFileType = ext[1:]
		}
	}
	switch *outputFileType {
	case "jpg", "jpeg", "png", "tif", "tiff":
		{
			params.SetCollageRenderer(AllCollageCreatorComponents["CollageRenderer_Raster"].(CollageCreator.CollageRenderer))
			ext := strings.ToLower(filepath.Ext(*outFile))
			if ext == "" {
				*outFile = *outFile + "." + *outputFileType
			}
		}
	case "svg":
		{
			params.SetCollageRenderer(AllCollageCreatorComponents["CollageRenderer_SVG"].(CollageCreator.CollageRenderer))
			ext := strings.ToLower(filepath.Ext(*outFile))
			if ext == "" {
				*outFile = *outFile + "." + *outputFileType
			}
		}
	case "sh":
		{
			params.SetCollageRenderer(AllCollageCreatorComponents["CollageRenderer_ImageMagickScript"].(CollageCreator.CollageRenderer))
			ext := strings.ToLower(filepath.Ext(*outFile))
			if ext == "" {
				*outFile = *outFile + "." + *outputFileType
			}
		}
	default:
		{
			log.Fatal("Unrecognized output type:", *outputFileType)
		}
	}

	params.SetInFiles(inputImages)
	params.SetOutFile(*outFile)
	aspectGeom, err := CollageCreator.ParseGeometry(*aspectRatio)
	if err != nil {
		log.Fatal("Error parsing aspect ratio:", err.Error())
	}
	params.SetAspectRatio(aspectGeom)
	paddingGeom, err := CollageCreator.ParseGeometry(*padding)
	if err != nil {
		log.Fatal("Error parsing padding:", err.Error())
	}
	params.SetPadding(paddingGeom)
	params.SetMinCanvasSize(CollageCreator.MustParseDims(*minCanvasSize))
	params.SetMaxCanvasSize(CollageCreator.MustParseDims(*maxCanvasSize))
	valid = params.ProgressMonitor().ParseCustomParameters(&params)
	valid = valid && params.InputImageReader().ParseCustomParameters(&params)
	valid = valid && params.DimensionInitializer().ParseCustomParameters(&params)
	valid = valid && params.PositionCalculator().ParseCustomParameters(&params)
	valid = valid && params.CollageRenderer().ParseCustomParameters(&params)
	if !valid {
		log.Fatal("Error parsing custom parameters")
	}

	errorlevel := CollageCreator.CreateCollage(&params)
	os.Exit(errorlevel)
}
