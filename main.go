package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// func copyfile using os package without io or ioutil
func copyFile(src, dst string) (err error) {
	in, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, in, 0644)
	if err != nil {
		return err
	}
	// set file modified date to the same as the original file
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	err = os.Chtimes(dst, fileInfo.ModTime(), fileInfo.ModTime())
	if err != nil {
		return err
	}
	return
}

func main() {
    savePath := os.Getenv("LOCALAPPDATA") + `\Packages\RyseupStudios.Roboquest_gdfnzxafmssey\SystemAppData\wgs`
    bkpPath := os.Getenv("USERPROFILE") + `\Desktop\Roboquest_saves_bkp`
    destPath := os.Getenv("LOCALAPPDATA") + `\Roboquest\Saved\SaveGames`

    // if destpath does not exist, create it
    if _, err := os.Stat(destPath); os.IsNotExist(err) {
        os.MkdirAll(destPath, os.ModePerm)
    }

    // if bkpPath does not exist, create it
    if _, err := os.Stat(bkpPath); os.IsNotExist(err) {
        fmt.Printf("Creating backup folder: %s\n", bkpPath)
        os.MkdirAll(bkpPath, os.ModePerm)
    }

    // Check if the file contains specific text to rename it
    checkFileType := func(filename string) string {
        fileContent, _ := ioutil.ReadFile(filename)
        fileContentStr := string(fileContent)
        //if file contains "Controls" then it is "Controls.sav" file
        switch {
        case strings.Contains(fileContentStr, "Controls"):
            return "Controls.sav"
        case strings.Contains(fileContentStr, "Graphics"):
            return "Graphics.sav"
        case strings.Contains(fileContentStr, "Localization"):
            return "Localization.sav"
        case strings.Contains(fileContentStr, "Settings"):
            return "Settings.sav"
        case strings.Contains(fileContentStr, "Profile"):
            return "Profile.sav"
        default:
            return ""
        }
    }

    // get list of folders inside savePath
    folders, _ := ioutil.ReadDir(savePath)

    // enter the folder not named "t"
    var folder os.FileInfo
    for _, f := range folders {
        if f.Name() != "t" && f.IsDir() {
            folder = f
            break
        }
    }

    folderPath := filepath.Join(savePath, folder.Name())

    // get list of files inside the folder
    subfolders, _ := ioutil.ReadDir(folderPath)
    var filesList [][]interface{}
    // open each folder and get the file that not contains "container" in the name
    for _, subfolder := range subfolders {
        subfolderPath := filepath.Join(folderPath, subfolder.Name())
        files, _ := ioutil.ReadDir(subfolderPath)
        for _, file := range files {
            if !file.IsDir() && !strings.Contains(file.Name(), "container") {
                filePath := filepath.Join(subfolderPath, file.Name())
                newFileName := checkFileType(filePath)
                if newFileName != "" {
                    // get file modified date
                    fileDate := file.ModTime()
                    filesList = append(filesList, []interface{}{filePath, newFileName, fileDate})
                } else {
                    fmt.Printf("File %s is not a valid save file\n", filePath)
                    os.Exit(1)
                }
            }
        }
    }

    var oldFile []interface{}
    // save position in the array of the file "Profile.sav" with the oldest date
    for _, file := range filesList {
        if strings.Contains(file[1].(string), "Profile.sav") {
            if oldFile != nil {
                if file[2].(time.Time).Before(oldFile[2].(time.Time)) {
                    oldFile = file
				}
			} else {
				oldFile = file
			}
		}
	}

	// find oldFile inside the array and rename [2] to ProfileCopy.sav
	for i, file := range filesList {
		if file[0] == oldFile[0] {
			filesList[i][1] = "ProfileCopy.sav"
		}
	}

	// copy files to destPath
	for _, file := range filesList {
		filePath := file[0].(string)
		// get old file name
		oldFileName := filepath.Base(filePath)
		newFileName := file[1].(string)
		newFilePath := filepath.Join(destPath, newFileName)
		fmt.Printf("Copying %s as %s\n", oldFileName, newFileName)
		err := copyFile(filePath, newFilePath)
		if err != nil {
			fmt.Println(err)
		}
		// copy old file to bkpPath
		bkpFilePath := filepath.Join(bkpPath, oldFileName)
		err = copyFile(filePath, bkpFilePath)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("Done! Press any key to exit...")
	fmt.Scanln()
}

