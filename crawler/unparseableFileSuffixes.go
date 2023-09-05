package crawler

// file suffixes that shoudln't/are not able to be parsed
var UnparseableFileSuffixes []string = []string{
	// file types
	".jpg",
	".jpeg",
	".png",
	".img",
	".svg",
	".gif",

	".pdf",
	".doc",
	".docx",
	".xls",
	".xlsx",
	".ppt",
	".pptx",

	".wav",
	".mp3",
	".mp4",
	".avi",
	".mov",

	".zip",
	".tar",
	".gz",
	".rar",
	".7z",

	".exe",
	".dll",
	".bin",
	".jar",

	".iso",
	".war",
	".ear",
	".class",
	".o",
	".obj",
	".so",
	".a",
	".lib",
	".rpm",
	".deb",
	".apk",
	".ipa",
	".dmg",
	".pkg",
	".app",
	".msi",

	".bmp",
	".tif",
	".tiff",
	".odt",
	".ods",
	".odp",
	".wmv",
	".flv",
	".sys",
	".ini",
	".bak",
	".tmp",
	".swp",
	".dat",
	".db",
	".sql",
	".rtf",
	".ott",
	".pot",
	".sldx",
	".sldm",
	".ppsx",
	".ppsm",
	".docm",
	".dotm",
	".xlsm",
	".xltx",
	".xltm",
	".xhtml",
	".xlam",
	".ppam",
	".docb",
	".mht",
	".mhtml",
	".eml",
	".msg",
	".oft",
	".vcf",
	".ics",
	".aar",
	".pyc",
	".pyd",
	".pdb",
	".dylib",
	".msp",
	".mst",
	".reg",

	".bat",
	".sh",
	".cmd",
	".ps1",
	".bash",
	".zsh",
	".csh",
	".tcsh",

	".ksh",
	".awk",
	".sed",

	".gpx",
	".kml",
	".kmz",
	".srt",
	".sub",
	".ass",
	".ssa",
	".vtt",
	".sbv",
	".mpsub",
	".lrc",
	".ttml",
	".dfxp",
	".smi",
	".xslt",
	".xsd",
	".wsdl",
	".soap",
	".protobuf",
	".thrift",
	".avro",
	".msgpack",
	".cbor",
	".pickle",
	".dill",
	".joblib",
	".hdf5",
	".pkl",
	".bz2",
	".xz",
	".lz4",
	".zstd",

	// files
	"wp-login.php",
	"wp-admin",
}
