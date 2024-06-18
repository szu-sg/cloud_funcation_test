package oss

// func Test_s3OSS_UploadObject(t *testing.T) {
// 	fmt.Println(os.Getwd())
// 	cl, err := NewS3OSS("xxx", "xxxx", "cn-north-1", "cloudide-component-cn-prod", "https://cloudide-component-cn-prod.bucket.vpce-0430d6f168157fef2-vko285as.s3.cn-north-1.vpce.amazonaws.com.cn")
// 	if err != nil {
// 		t.Errorf("NewS3OSS() error = %v", err)
// 	}
// 	ctx := context.Background()
// 	p := "blank.tar.gz"
// 	fileInfo, err := os.Stat(p)
// 	if err != nil {
// 		t.Errorf("os.Stat() error = %v", err)
// 	}
// 	file, err := os.Open(p)
// 	if err != nil {
// 		t.Errorf("os.Open() error = %v", err)
// 	}
// 	defer file.Close()

// 	// 将文件转换为 io.Reader
// 	reader := io.Reader(file)
// 	u, err := cl.UploadObject(ctx, "test/oss/blank.tar.gz", reader, fileInfo.Size())
// 	if err != nil {
// 		t.Errorf("s3OSS.UploadObject() error = %v", err)
// 	}
// 	fmt.Println(u)
// }
