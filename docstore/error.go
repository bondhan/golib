package docstore

type DocstoreError string

func (e DocstoreError) Error() string { return string(e) }

const NotFound = DocstoreError("[docstore] document not found")
const EndOfDoc = DocstoreError("[docstore] end of documents")
const NothingUpdated = DocstoreError("[docstore] nothing updated")
const OperationNotSupported = DocstoreError("[docstore] operation not supported")
