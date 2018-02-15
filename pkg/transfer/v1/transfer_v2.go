package transfer

// //Database keeps metadata and transactional state of a dataset
// type Database interface {
// }
//
// //Objects facilitates long term blob storage for large files
// type Objects interface {
// 	Get(key string, w io.WriterAt) error
// 	Put(key string, r io.Reader) error
// 	Delete(key string) error
// }
//
// //Manager sets up handles
// type Manager struct {
// 	db  Database
// 	obj Objects
// }
//
// //NewManager will setup a dataset transfer manager
// func NewManager(db Database, obj Objects) (m *Manager, err error) {
// 	return m, nil
// }
//
// //Open will claim a dataset (transfer) handle
// func (m *Manager) Open(dataset string) (s *Handle, err error) {
// 	return s, nil
// }
//
// //Handle represents a claim on the access to a dataset
// type Handle struct{}
//
// //Upload 'path' to the dataset of this handle
// func (s *Handle) Upload(path string, progress chan<- struct{}) error {
//
// 	//@TODO periodically send keep-alive call
//
// 	return nil
// }
//
// //Download the contents of this handles's dataset to 'path'
// func (s *Handle) Download(path string, progress chan<- struct{}) error {
// 	return nil
// }
//
// //Close will release the claim on a dataset
// func (s *Handle) Close() error {
// 	return nil
// }
//
// //KubeDatabase uses a Kubernetes custom resource to manage dataset state
// type KubeDatabase struct {
// 	namespace string
// }
//
// //Open a new dataset handle
// func (db *KubeDatabase) Open(datasetID string) (h Handle, err error) {
// 	return h, nil
// }
