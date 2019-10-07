package cblcgo
// /*
// #cgo LDFLAGS: -L. -lCouchbaseLiteC
// #include <stdlib.h>
// #include <stdio.h>
// //#include "include/CBLLog.h"
// #include "include/CBLLog.h"

// */
// import "C"

// type LogLevel uint8

// const (
// 	LogDebug LogLevel = iota
//     LogVerbose
//     LogInfo
//     LogWarning
//     LogError
//     LogNone
// )

// type LogDomain uint8

// const (
// 	LogDomainAll LogDomain = iota
//     LogDomainDatabase
//     LogDomainQuery
//     LogDomainReplicator
//     LogDomainNetwork
// )

// /** An object containing properties for file logging configuration 
//     @warning \ref usePlaintext results in significantly larger log files; we recommend turning
// 	it off in production. */
// 	// typedef struct {
// 	// 	const char* directory;          ///< The directory to write logs to (UTF-8 encoded)
// 	// 	const uint32_t maxRotateCount;  ///< The maximum number of *rotated* logs to keep (i.e. the total number of logs will be one more)
// 	// 	const size_t maxSize;           ///< The max size to write to a log file before rotating (best-effort)
// 	// 	const bool usePlaintext;        ///< Whether or not to log in plaintext (as opposed to binary)
// 	// } CBLLogFileConfiguration;
// type LogFileConfiguration struct {
// 	Directory string ///< The directory to write logs to (UTF-8 encoded)
// 	MaxRotateCount uint32 ///< The maximum number of *rotated* logs to keep (i.e. the total number of logs will be one more)
// 	MaxSize uint64 ///< The max size to write to a log file before rotating (best-effort)
// 	UsePlainText bool ///< Whether or not to log in plaintext (as opposed to binary)
// }
	
// /** A callback function for handling log messages
// 	@param  level The level of the message being received
// 	@param  domain The domain of the message being received
// 	@param  message The message being received (UTF-8 encoded) */
// //typedef void(*CBLLogCallback)(CBLLogLevel level, CBLLogDomain domain, const char* message);
// type LogCallback func (level LogLevel, domain LogDomain, message string)

// /** Gets the current log level for debug console logging */
// // CBLLogLevel CBLLog_ConsoleLevel();
// // func GetLogLevel() LogLevel {
// // 	lvl := C.CBLLog_ConsoleLevel()
// // 	return LogLevel(lvl)
// // }

// /** Sets the debug console log level */
// //void CBLLog_SetConsoleLevel(CBLLogLevel);
// func SetLogLevel(level LogLevel) {
// 	C.CBLLog_SetConsoleLevel(CBLLogLevel(level))
// }

// // /** Gets the current file logging config */
// // // const CBLLogFileConfiguration* CBLLog_FileConfig();
// // func LogFileConfig() *LogFileConfiguration {
// // 	c_config := C.CBLLog_FileConfig()
// // 	dir := C.GoString(c_config.directory)
// // 	config := LogFileConfiguration{dir, uint32(c_config.maxRotateCount), uint64(c_config.maxSize), bool(c_config.usePlaintext)}
// // 	return &config
// // }

// // /** Sets the file logging configuration */
// // // void CBLLog_SetFileConfig(CBLLogFileConfiguration);
// // func SetLogFileConfig(config LogFileConfiguration) {
// // 	c_config := C.CBLLogFileConfiguration{}
// // 	c_config.directory = C.CString(config.Directory)
// // 	c_config.maxRotateCount = C.uint32_t(config.MaxRotateCount)
// // 	c_config.maxSize = C.size_t(config.MaxSize)
// // 	c_config.usePlaintext = C.bool(config.UsePlainText)
// // 	C.CBLLog_SetFileConfig(c_config)
// // }

// // /** Gets the current log callback */
// // // CBLLogCallback CBLLog_Callback();
// // func GetLogCallback() LogCallback {
// // 	callback := C.CBLLog_Callback()
// // 	return LogCallback(callback)
// // }

// // /** Sets the callback for receiving log messages */
// // // void CBLLog_SetCallback(CBLLogCallback);
// // func SetLogCallback(callback LogCallback) {
// // 	C.CBLLog_SetCallback(C.CBLLogCallback(callback))
// // }
	
// /** @} */