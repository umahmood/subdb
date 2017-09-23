# SubDB

A Go library to access SubDB. A free, centralized subtitle database intended to 
be used only by opensource and non-commercial softwares.

# Installation

> $ go get github.com/umahmood/subdb

# Usage

Setting the correct user agent string is a requirement to access SubDB.
```
package main

import (
    "fmt"
    
    "github.com/umahmood/subdb"
)

func main() {
    subdb := &subdb.API{}
    subdb.SetUserAgent("Acme", "1.0", "https://acme.org")
}
```

List the languages of all subtitles stored on the SubDB database
```
    langs, err := subdb.Languages()
    if err != nil {
     fmt.Println(err)
     return
    }
    fmt.Println(langs)
```

Output:
```
[en es fr it nl pl pt ro sv tr]
```

Searching subtitles:
```
    fileName := "/home/fry/movies/Avatar.mp4"
    subs, err := subdb.Search(fileName)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(subs)
```

Output:
```
[en:2 fr:1]
```

Downloading subtitles:
```
    subtitles, err := subdb.Download(fileName, "en")
    if err != nil {
     fmt.Println(err)
     return
    }
    fmt.Println(subtitles)
```

Output:
```
1
00:00:01,280 --> 00:00:06,000
<i>Sooner or later though,
you always have to wake up.</i>

2
00:00:18,000 --> 00:00:20,685
<i>In cryo you don't dream at all.</i>

...

```
Uploading subtitles:
```
    subtitleFile := "/home/fry/movies/Avatar.srt"
    err := subdb.Upload(fileName, subtitleFile)
    if err != nil {
     fmt.Println(err)
    }
}
```

# Documentation

> http://godoc.org/github.com/umahmood/subdb

# License

See the [LICENSE](LICENSE.md) file for license rights and limitations (MIT).
