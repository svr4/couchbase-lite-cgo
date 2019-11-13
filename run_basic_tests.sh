tests=(TestConnection TestSaveAndDeleteDocuments TestSaveAndRetrieveDocuments TestProperties TestDocumentListener TestQuery TestBlob TestListeners TestNotificationCallback)
for i in ${tests[@]}; do
    ./cblcgo.test -test.v -test.run $i
    echo $i
done
echo "Tests Done"
