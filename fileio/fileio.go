package fileio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	lz4 "github.com/pierrec/lz4/v4"
)

type MapHeaderInput struct {
	Version1           uint32
	Version2           uint32
	TotalActions       uint16
	CurrentTurn        uint32
	CurrentPlayerIndex uint8
	MaxUnitId          uint32
	CurrentGameState   uint8
	Seed               int32
	TurnLimit          uint32
	ScoreLimit         uint32
	WinByCapital       uint8
	UnknownSettings    [6]byte
	GameModeBase       uint8
	GameModeRules      uint8
}

type MapHeaderOutput struct {
	MapHeaderInput       MapHeaderInput
	MapName              string
	MapSquareSize        int
	DisabledTribesArr    []int
	UnlockedTribesArr    []int
	GameDifficulty       int
	NumOpponents         int
	GameType             int
	MapPreset            int
	TurnTimeLimitMinutes int
	UnknownFloat1        float32
	UnknownFloat2        float32
	BaseTimeSeconds      float32
	TimeSettings         []int
	SelectedTribeSkins   []TribeSkin
	MapWidth             int
	MapHeight            int
}

type TribeSkin struct {
	Tribe int
	Skin  int
}

type TileDataHeader struct {
	WorldCoordinates   [2]uint32
	Terrain            uint16
	Climate            uint16
	Altitude           int16
	Owner              uint8
	Capital            uint8
	CapitalCoordinates [2]int32
}

type TileData struct {
	WorldCoordinates           [2]int
	Terrain                    int
	Climate                    int
	Altitude                   int
	Owner                      int
	Capital                    int
	CapitalCoordinates         [2]int
	ResourceExists             bool
	ResourceType               int
	ImprovementExists          bool
	ImprovementType            int
	ImprovementData            *ImprovementData
	Unit                       *UnitData
	PassengerUnit              *UnitData
	UnitEffectData             []int // flags: 0 - ice, 1 - poison, 2 - boost, 3 - invisible
	UnitDirectionData          []int // contains direction flag (0 - southwest, 1 - west, 2 - northwest, 3 - north, 4 - northeast, 5 - east, 6 - southwest, 7 - south)
	PassengerUnitEffectData    []int
	PassengerUnitDirectionData []int
	PlayerVisibility           []int
	HasRoad                    bool
	HasWaterRoute              bool
	TileSkin                   int
	Unknown                    []int
}

type ActionBuild struct {
	PlayerId        uint8
	ImprovementType uint16
	Coordinates     [2]uint32
}

type ActionAttack struct {
	PlayerId uint8
	UnitId   uint32
	Origin   [2]uint32
	Target   [2]uint32
}

type ActionRecover struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

type ActionTrain struct {
	PlayerId uint8
	UnitType uint16
	Position [2]uint32
}

type ActionMove struct {
	PlayerId    uint8
	OldPosition [2]uint32
	NewPosition [2]uint32
	UnitId      uint32
}

type ActionCaptureCity struct {
	PlayerId    uint8
	UnitId      uint32
	Coordinates [2]uint32
}

type ActionResearch struct {
	PlayerId uint8
	TechType uint16
}

type ActionDestroyImprovement struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

type ActionCityReward struct {
	PlayerId    uint8
	Coordinates [2]uint32
	Reward      uint16
}

type ActionPromote struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

type ActionExamineRuins struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

type ActionEndTurn struct {
	PlayerId uint8
}

type ActionUpgrade struct {
	PlayerId    uint8
	UnitType    uint16
	Coordinates [2]uint32
}

type ActionCityLevelUp struct {
	PlayerId    uint8
	Coordinates [2]uint32
}

type PlayerData struct {
	PlayerId             int
	Name                 string
	AccountId            string
	AutoPlay             bool
	StartTileCoordinates [2]int
	Tribe                int
	UnknownByte1         int
	DifficultyHandicap   int
	AggressionsByPlayers []PlayerAggression
	Currency             int
	Score                int
	UnknownInt2          int
	NumCities            int
	AvailableTech        []int
	EncounteredPlayers   []int
	Tasks                []PlayerTaskData
	TotalUnitsKilled     int
	TotalUnitsLost       int
	TotalTribesDestroyed int
	OverrideColor        []int
	OverrideTribe        byte
	UniqueImprovements   []int
	DiplomacyArr         []DiplomacyData
	DiplomacyMessages    []DiplomacyMessage
	DestroyedByTribe     int
	DestroyedTurn        int
	UnknownBuffer2       []int
	EndScore             int
	PlayerSkin           int
	UnknownBuffer3       []int
}

type PlayerAggression struct {
	PlayerId   int
	Aggression int
}

type UnitData struct {
	Id                 uint32
	Owner              uint8
	UnitType           uint16
	FollowerUnitId     uint32 // only initialized for cymanti centipedes and segments
	LeaderUnitId       uint32 // only initialized for cymanti centipedes and segments
	CurrentCoordinates [2]int32
	HomeCoordinates    [2]int32
	Health             uint16 // should be divided by 10 to get value ingame
	PromotionLevel     uint16
	Experience         uint16
	Moved              bool
	Attacked           bool
	Flipped            bool
	CreatedTurn        uint16
}

type ImprovementData struct {
	Level                  int
	FoundedTurn            int
	CurrentPopulation      int
	TotalPopulation        int
	Production             int
	BaseScore              int
	BorderSize             int // For cities, 1 is default, 2 is expanded border
	UpgradeCount           int // For cities, seems to be -1 * (level - 1). Level 1 is starting point and no upgrades.
	ConnectedPlayerCapital int
	HasCityName            int
	CityName               string
	FoundedTribe           int
	CityRewards            []int
	RebellionFlag          int
	RebellionBuffer        []int
}

type PlayerTaskData struct {
	Type   int
	Buffer []int
}

type DiplomacyMessage struct {
	MessageType int
	Sender      int
}

type DiplomacyData struct {
	PlayerId               uint8
	DiplomacyRelationState uint8
	LastAttackTurn         int32
	EmbassyLevel           uint8
	LastPeaceBrokenTurn    int32
	FirstMeet              int32
	EmbassyBuildTurn       int32
	PreviousAttackTurn     int32
}

type PolytopiaSaveOutput struct {
	MapHeight       int
	MapWidth        int
	PlayerData      []PlayerData
	OwnerTribeMap   map[int]int
	InitialTileData [][]TileData
	TileData        [][]TileData
	MaxTurn         int
	TurnCaptureMap  map[int][]ActionCaptureCity
}

func readVarString(reader *io.SectionReader, varName string) string {
	variableLength := uint8(0)
	if err := binary.Read(reader, binary.LittleEndian, &variableLength); err != nil {
		log.Fatal("Failed to load variable length: ", err)
	}

	stringValue := make([]byte, variableLength)
	if err := binary.Read(reader, binary.LittleEndian, &stringValue); err != nil {
		log.Fatal(fmt.Sprintf("Failed to load string value. Variable length: %v, name: %s. Error:", variableLength, varName), err)
	}

	return string(stringValue[:])
}

func unsafeReadUint32(reader *io.SectionReader) uint32 {
	unsignedIntValue := uint32(0)
	if err := binary.Read(reader, binary.LittleEndian, &unsignedIntValue); err != nil {
		log.Fatal("Failed to load uint32: ", err)
	}
	return unsignedIntValue
}

func unsafeReadInt32(reader *io.SectionReader) int32 {
	signedIntValue := int32(0)
	if err := binary.Read(reader, binary.LittleEndian, &signedIntValue); err != nil {
		log.Fatal("Failed to load int32: ", err)
	}
	return signedIntValue
}

func unsafeReadUint16(reader *io.SectionReader) uint16 {
	unsignedIntValue := uint16(0)
	if err := binary.Read(reader, binary.LittleEndian, &unsignedIntValue); err != nil {
		log.Fatal("Failed to load uint16: ", err)
	}
	return unsignedIntValue
}

func unsafeReadInt16(reader *io.SectionReader) int16 {
	signedIntValue := int16(0)
	if err := binary.Read(reader, binary.LittleEndian, &signedIntValue); err != nil {
		log.Fatal("Failed to load int16: ", err)
	}
	return signedIntValue
}

func unsafeReadUint8(reader *io.SectionReader) uint8 {
	unsignedIntValue := uint8(0)
	if err := binary.Read(reader, binary.LittleEndian, &unsignedIntValue); err != nil {
		log.Fatal("Failed to load uint8: ", err)
	}
	return unsignedIntValue
}

func unsafeReadInt8(reader *io.SectionReader) int8 {
	signedIntValue := int8(0)
	if err := binary.Read(reader, binary.LittleEndian, &signedIntValue); err != nil {
		log.Fatal("Failed to load int8: ", err)
	}
	return signedIntValue
}

func unsafeReadFloat32(reader *io.SectionReader) float32 {
	floatValue := float32(0)
	if err := binary.Read(reader, binary.LittleEndian, &floatValue); err != nil {
		log.Fatal("Failed to load float32: ", err)
	}
	return floatValue
}

func readFixedList(streamReader *io.SectionReader, listSize int) []byte {
	buffer := make([]byte, listSize)
	if err := binary.Read(streamReader, binary.LittleEndian, &buffer); err != nil {
		log.Fatal("Failed to load buffer: ", err)
	}
	return buffer
}

func convertByteListToInt(oldArr []byte) []int {
	newArr := make([]int, len(oldArr))
	for i := 0; i < len(newArr); i++ {
		newArr[i] = int(oldArr[i])
	}
	return newArr
}

func buildReaderForDecompressedFile(inputFilename string) (*bytes.Reader, int) {
	inputFile, err := os.Open(inputFilename)
	defer inputFile.Close()
	if err != nil {
		log.Fatal("Failed to load state file: ", err)
		return nil, 0
	}

	inputBuffer := new(bytes.Buffer)
	inputBuffer.ReadFrom(inputFile)

	inputBytes := inputBuffer.Bytes()
	firstByte := inputBytes[0]
	sizeOfDiff := ((firstByte >> 6) & 3)
	if sizeOfDiff == 3 {
		sizeOfDiff = 4
	}
	dataOffset := 1 + int(sizeOfDiff)
	var resultDiff int
	if sizeOfDiff == 4 {
		resultDiff = int(binary.LittleEndian.Uint32(inputBytes[1 : 1+sizeOfDiff]))
	} else if sizeOfDiff == 2 {
		resultDiff = int(binary.LittleEndian.Uint16(inputBytes[1 : 1+sizeOfDiff]))
	} else {
		log.Fatal("Header sizeOfDiff is unrecognized value: ", sizeOfDiff)
	}
	dataLength := len(inputBytes) - dataOffset
	resultLength := dataLength + resultDiff

	// decompress
	decompressedContents := make([]byte, resultLength)
	decompressedLength, err := lz4.UncompressBlock(inputBytes[dataOffset:], decompressedContents)
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(decompressedContents), decompressedLength
}

func DeserializeImprovementDataFromBytes(streamReader *io.SectionReader) ImprovementData {
	level := unsafeReadUint16(streamReader)
	foundedTurn := unsafeReadUint16(streamReader)
	currentPopulation := unsafeReadInt16(streamReader)
	totalPopulation := unsafeReadUint16(streamReader)
	production := unsafeReadInt16(streamReader)
	baseScore := unsafeReadInt16(streamReader)
	borderSize := unsafeReadInt16(streamReader)
	upgradeCount := unsafeReadInt16(streamReader)
	connectedPlayerCapital := unsafeReadUint8(streamReader)
	hasCityName := unsafeReadUint8(streamReader)
	cityName := ""
	if hasCityName == 1 {
		cityName = readVarString(streamReader, "CityName")
	}

	foundedTribe := unsafeReadUint8(streamReader)
	cityRewardsSize := unsafeReadUint16(streamReader)
	cityRewards := make([]int, cityRewardsSize)
	for i := 0; i < int(cityRewardsSize); i++ {
		cityReward := unsafeReadUint16(streamReader)
		cityRewards[i] = int(cityReward)
	}

	rebellionFlag := unsafeReadUint16(streamReader)
	rebellionBuffer := []int{}
	if rebellionFlag != 0 {
		rebellionBuffer = convertByteListToInt(readFixedList(streamReader, 2))
	}

	return ImprovementData{
		Level:                  int(level),
		FoundedTurn:            int(foundedTurn),
		CurrentPopulation:      int(currentPopulation),
		TotalPopulation:        int(totalPopulation),
		Production:             int(production),
		BaseScore:              int(baseScore),
		BorderSize:             int(borderSize),
		UpgradeCount:           int(upgradeCount),
		ConnectedPlayerCapital: int(connectedPlayerCapital),
		HasCityName:            int(hasCityName),
		CityName:               cityName,
		FoundedTribe:           int(foundedTribe),
		CityRewards:            cityRewards,
		RebellionFlag:          int(rebellionFlag),
		RebellionBuffer:        rebellionBuffer,
	}
}

func readTileData(streamReader *io.SectionReader, tileData [][]TileData, mapWidth int, mapHeight int) {
	allUnitData := make([]UnitData, 0)

	for i := 0; i < int(mapHeight); i++ {
		for j := 0; j < int(mapWidth); j++ {
			tileDataHeader := TileDataHeader{}
			if err := binary.Read(streamReader, binary.LittleEndian, &tileDataHeader); err != nil {
				log.Fatal("Failed to load tileDataHeader: ", err)
			}

			if int(tileDataHeader.WorldCoordinates[0]) != j || int(tileDataHeader.WorldCoordinates[1]) != i {
				log.Fatal(fmt.Sprintf("File reached unexpected location. Iteration (%v, %v) isn't equal to world coordinates (%v, %v)",
					i, j, tileDataHeader.WorldCoordinates[0], tileDataHeader.WorldCoordinates[1]))
			}

			resourceExistsFlag := unsafeReadUint8(streamReader)
			resourceType := -1
			if resourceExistsFlag == 1 {
				resourceType = int(unsafeReadUint16(streamReader))
			}

			improvementExistsFlag := unsafeReadUint8(streamReader)
			improvementType := -1
			if improvementExistsFlag == 1 {
				improvementType = int(unsafeReadUint16(streamReader))
			}

			var improvementDataPtr *ImprovementData
			if improvementType != -1 {
				improvementData := DeserializeImprovementDataFromBytes(streamReader)
				improvementDataPtr = &improvementData
			}

			// Read unit data
			hasUnitFlag := unsafeReadUint8(streamReader)
			var unitDataPtr *UnitData
			var passengerUnitDataPtr *UnitData

			unitEffectData := make([]int, 0)
			unitDirectionData := make([]byte, 0)
			passengerUnitEffectData := make([]int, 0)
			passengerUnitDirectionData := make([]byte, 0)
			if hasUnitFlag == 1 {
				unitData := UnitData{}
				if err := binary.Read(streamReader, binary.LittleEndian, &unitData); err != nil {
					log.Fatal("Failed to load buffer: ", err)
				}
				unitDataPtr = &unitData

				hasOtherUnitFlag := unsafeReadUint8(streamReader)
				if hasOtherUnitFlag == 1 {
					// If unit embarks or disembarks, a new unit is created in the backend, but it's still the same unit in the game
					passengerUnitData := UnitData{}
					if err := binary.Read(streamReader, binary.LittleEndian, &passengerUnitData); err != nil {
						log.Fatal("Failed to load buffer: ", err)
					}
					passengerUnitDataPtr = &passengerUnitData

					// might be other unit flag for passenger unit
					// should always be zero because passenger unit can't carry another unit
					unknownFlag := int(unsafeReadUint8(streamReader))
					if unknownFlag != 0 {
						log.Fatal("Passenger unit's other unit flag isn't zero")
					}

					passengerUnitEffectCount := int(unsafeReadUint16(streamReader))
					passengerUnitEffectData = make([]int, 0)
					for statusIndex := 0; statusIndex < passengerUnitEffectCount; statusIndex++ {
						passengerUnitEffectData = append(passengerUnitEffectData, int(unsafeReadUint16(streamReader)))
					}
					passengerUnitDirectionData = readFixedList(streamReader, 5)

					unitEffectCount := int(unsafeReadUint16(streamReader))
					unitEffectData = make([]int, 0)
					for statusIndex := 0; statusIndex < unitEffectCount; statusIndex++ {
						unitEffectData = append(unitEffectData, int(unsafeReadUint16(streamReader)))
					}
					unitDirectionData = readFixedList(streamReader, 5)
				} else {
					unitEffectCount := int(unsafeReadUint16(streamReader))
					unitEffectData = make([]int, 0)
					for statusIndex := 0; statusIndex < unitEffectCount; statusIndex++ {
						unitEffectData = append(unitEffectData, int(unsafeReadUint16(streamReader)))
					}
					unitDirectionData = readFixedList(streamReader, 5)
				}
			}

			playerVisibilityListSize := unsafeReadUint8(streamReader)
			playerVisibilityList := convertByteListToInt(readFixedList(streamReader, int(playerVisibilityListSize)))
			hasRoad := unsafeReadUint8(streamReader)
			hasWaterRoute := unsafeReadUint8(streamReader)
			tileSkin := unsafeReadUint16(streamReader)
			unknown := convertByteListToInt(readFixedList(streamReader, 2))

			tileData[i][j] = TileData{
				WorldCoordinates:           [2]int{int(tileDataHeader.WorldCoordinates[0]), int(tileDataHeader.WorldCoordinates[1])},
				Terrain:                    int(tileDataHeader.Terrain),
				Climate:                    int(tileDataHeader.Climate),
				Altitude:                   int(tileDataHeader.Altitude),
				Owner:                      int(tileDataHeader.Owner),
				Capital:                    int(tileDataHeader.Capital),
				CapitalCoordinates:         [2]int{int(tileDataHeader.CapitalCoordinates[0]), int(tileDataHeader.CapitalCoordinates[1])},
				ResourceExists:             resourceExistsFlag != 0,
				ResourceType:               resourceType,
				ImprovementExists:          improvementExistsFlag != 0,
				ImprovementType:            improvementType,
				ImprovementData:            improvementDataPtr,
				Unit:                       unitDataPtr,
				PassengerUnit:              passengerUnitDataPtr,
				UnitEffectData:             unitEffectData,
				UnitDirectionData:          convertByteListToInt(unitDirectionData),
				PassengerUnitEffectData:    passengerUnitEffectData,
				PassengerUnitDirectionData: convertByteListToInt(passengerUnitDirectionData),
				PlayerVisibility:           playerVisibilityList,
				HasRoad:                    hasRoad != 0,
				HasWaterRoute:              hasWaterRoute != 0,
				TileSkin:                   int(tileSkin),
				Unknown:                    unknown,
			}

			if tileData[i][j].Unit != nil {
				allUnitData = append(allUnitData, *tileData[i][j].Unit)
			}
		}
	}
}

func DeserializeMapHeaderFromBytes(streamReader *io.SectionReader) MapHeaderOutput {
	mapHeaderInput := MapHeaderInput{}
	if err := binary.Read(streamReader, binary.LittleEndian, &mapHeaderInput); err != nil {
		log.Fatal("Failed to load MapHeaderInput: ", err)
	}

	mapName := readVarString(streamReader, "MapName")

	// map dimenions is a square: squareSize x squareSize
	squareSize := int(unsafeReadUint32(streamReader))

	disabledTribesSize := unsafeReadUint16(streamReader)
	disabledTribesArr := make([]int, disabledTribesSize)
	for i := 0; i < int(disabledTribesSize); i++ {
		disabledTribesArr[i] = int(unsafeReadUint16(streamReader))
	}

	unlockedTribesSize := unsafeReadUint16(streamReader)
	unlockedTribesArr := make([]int, unlockedTribesSize)
	for i := 0; i < int(unlockedTribesSize); i++ {
		unlockedTribesArr[i] = int(unsafeReadUint16(streamReader))
	}

	gameDifficulty := unsafeReadUint16(streamReader)
	numOpponents := unsafeReadUint32(streamReader)
	gameType := unsafeReadUint16(streamReader)
	mapPreset := unsafeReadUint8(streamReader)
	turnTimeLimitMinutes := unsafeReadInt32(streamReader)
	unknownFloat1 := unsafeReadFloat32(streamReader)
	unknownFloat2 := unsafeReadFloat32(streamReader)
	baseTimeSeconds := unsafeReadFloat32(streamReader)
	timeSettings := readFixedList(streamReader, 4)

	selectedTribeSkinSize := unsafeReadUint32(streamReader)
	selectedTribeSkins := make([]TribeSkin, int(selectedTribeSkinSize))
	for i := 0; i < int(selectedTribeSkinSize); i++ {
		tribe := unsafeReadUint16(streamReader)
		skin := unsafeReadUint16(streamReader)
		selectedTribeSkins[i] = TribeSkin{
			Tribe: int(tribe),
			Skin:  int(skin),
		}
	}

	mapWidth := unsafeReadUint16(streamReader)
	mapHeight := unsafeReadUint16(streamReader)

	return MapHeaderOutput{
		MapHeaderInput:       mapHeaderInput,
		MapName:              mapName,
		MapSquareSize:        squareSize,
		DisabledTribesArr:    disabledTribesArr,
		UnlockedTribesArr:    unlockedTribesArr,
		GameDifficulty:       int(gameDifficulty),
		NumOpponents:         int(numOpponents),
		GameType:             int(gameType),
		MapPreset:            int(mapPreset),
		TurnTimeLimitMinutes: int(turnTimeLimitMinutes),
		UnknownFloat1:        unknownFloat1,
		UnknownFloat2:        unknownFloat2,
		BaseTimeSeconds:      baseTimeSeconds,
		TimeSettings:         convertByteListToInt(timeSettings),
		SelectedTribeSkins:   selectedTribeSkins,
		MapWidth:             int(mapWidth),
		MapHeight:            int(mapHeight),
	}
}

func DeserializePlayerDataFromBytes(streamReader *io.SectionReader) PlayerData {
	playerId := unsafeReadUint8(streamReader)
	playerName := readVarString(streamReader, "playerName")
	playerAccountId := readVarString(streamReader, "playerAccountId")
	autoPlay := unsafeReadUint8(streamReader)
	startTileCoordinates1 := unsafeReadInt32(streamReader)
	startTileCoordinates2 := unsafeReadInt32(streamReader)
	tribe := unsafeReadUint16(streamReader)
	unknownByte1 := unsafeReadUint8(streamReader)
	difficultyHandicap := unsafeReadUint32(streamReader)

	aggressionArrLen := unsafeReadUint16(streamReader)
	aggressionsByPlayers := make([]PlayerAggression, 0)
	for i := 0; i < int(aggressionArrLen); i++ {
		playerIdOther := unsafeReadUint8(streamReader)
		aggression := unsafeReadInt32(streamReader)
		aggressionsByPlayers = append(aggressionsByPlayers, PlayerAggression{
			PlayerId:   int(playerIdOther),
			Aggression: int(aggression),
		})
	}

	currency := unsafeReadUint32(streamReader)
	score := unsafeReadUint32(streamReader)
	unknownInt2 := unsafeReadUint32(streamReader)
	numCities := unsafeReadUint16(streamReader)

	techArrayLen := unsafeReadUint16(streamReader)
	techArray := make([]int, techArrayLen)
	for i := 0; i < int(techArrayLen); i++ {
		techType := unsafeReadUint16(streamReader)
		techArray[i] = int(techType)
	}

	encounteredPlayersLen := unsafeReadUint16(streamReader)
	encounteredPlayers := make([]int, 0)
	for i := 0; i < int(encounteredPlayersLen); i++ {
		playerId := unsafeReadUint8(streamReader)
		encounteredPlayers = append(encounteredPlayers, int(playerId))
	}

	numTasks := unsafeReadInt16(streamReader)
	taskArr := make([]PlayerTaskData, int(numTasks))
	for i := 0; i < int(numTasks); i++ {
		taskType := unsafeReadInt16(streamReader)

		var buffer []byte
		if taskType == 1 || taskType == 5 { // Task type 1 is Pacifist, type 5 is Killer
			buffer = readFixedList(streamReader, 6) // Extra buffer contains a uint32
		} else if taskType >= 1 && taskType <= 8 {
			buffer = readFixedList(streamReader, 2)
		} else {
			log.Fatal("Invalid task type:", taskType)
		}
		taskArr[i] = PlayerTaskData{
			Type:   int(taskType),
			Buffer: convertByteListToInt(buffer),
		}
	}

	totalKills := unsafeReadInt32(streamReader)
	totalLosses := unsafeReadInt32(streamReader)
	totalTribesDestroyed := unsafeReadInt32(streamReader)
	overrideColor := convertByteListToInt(readFixedList(streamReader, 4))
	overrideTribe := unsafeReadUint8(streamReader)

	playerUniqueImprovementsSize := unsafeReadUint16(streamReader)
	playerUniqueImprovements := make([]int, int(playerUniqueImprovementsSize))
	for i := 0; i < int(playerUniqueImprovementsSize); i++ {
		improvement := unsafeReadUint16(streamReader)
		playerUniqueImprovements[i] = int(improvement)
	}

	diplomacyArrLen := unsafeReadUint16(streamReader)
	diplomacyArr := make([]DiplomacyData, int(diplomacyArrLen))
	for i := 0; i < len(diplomacyArr); i++ {
		diplomacyData := DiplomacyData{}
		if err := binary.Read(streamReader, binary.LittleEndian, &diplomacyData); err != nil {
			log.Fatal("Failed to load diplomacyData: ", err)
		}
		diplomacyArr[i] = diplomacyData
	}

	diplomacyMessagesSize := unsafeReadUint16(streamReader)
	diplomacyMessagesArr := make([]DiplomacyMessage, int(diplomacyMessagesSize))
	for i := 0; i < int(diplomacyMessagesSize); i++ {
		messageType := unsafeReadUint8(streamReader)
		sender := unsafeReadUint8(streamReader)

		diplomacyMessagesArr[i] = DiplomacyMessage{
			MessageType: int(messageType),
			Sender:      int(sender),
		}
	}

	destroyedByTribe := unsafeReadUint8(streamReader)
	destroyedTurn := unsafeReadUint32(streamReader)
	unknownBuffer2 := convertByteListToInt(readFixedList(streamReader, 4))
	endScore := unsafeReadInt32(streamReader)
	playerSkin := unsafeReadUint16(streamReader)
	unknownBuffer3 := convertByteListToInt(readFixedList(streamReader, 4))

	return PlayerData{
		PlayerId:             int(playerId),
		Name:                 playerName,
		AccountId:            playerAccountId,
		AutoPlay:             int(autoPlay) != 0,
		StartTileCoordinates: [2]int{int(startTileCoordinates1), int(startTileCoordinates2)},
		Tribe:                int(tribe),
		UnknownByte1:         int(unknownByte1),
		DifficultyHandicap:   int(difficultyHandicap),
		AggressionsByPlayers: aggressionsByPlayers,
		Currency:             int(currency),
		Score:                int(score),
		UnknownInt2:          int(unknownInt2),
		NumCities:            int(numCities),
		AvailableTech:        techArray,
		EncounteredPlayers:   encounteredPlayers,
		Tasks:                taskArr,
		TotalUnitsKilled:     int(totalKills),
		TotalUnitsLost:       int(totalLosses),
		TotalTribesDestroyed: int(totalTribesDestroyed),
		OverrideColor:        overrideColor,
		OverrideTribe:        overrideTribe,
		UniqueImprovements:   playerUniqueImprovements,
		DiplomacyArr:         diplomacyArr,
		DiplomacyMessages:    diplomacyMessagesArr,
		DestroyedByTribe:     int(destroyedByTribe),
		DestroyedTurn:        int(destroyedTurn),
		UnknownBuffer2:       unknownBuffer2,
		EndScore:             int(endScore),
		PlayerSkin:           int(playerSkin),
		UnknownBuffer3:       unknownBuffer3,
	}
}

func readAllPlayerData(streamReader *io.SectionReader) []PlayerData {
	numPlayers := unsafeReadUint16(streamReader)
	allPlayerData := make([]PlayerData, int(numPlayers))

	for i := 0; i < int(numPlayers); i++ {
		playerData := DeserializePlayerDataFromBytes(streamReader)
		allPlayerData[i] = playerData
	}
	return allPlayerData
}

func buildOwnerTribeMap(allPlayerData []PlayerData) map[int]int {
	ownerTribeMap := make(map[int]int)

	for i := 0; i < len(allPlayerData); i++ {
		playerData := allPlayerData[i]
		mappedTribe, ok := ownerTribeMap[playerData.PlayerId]
		if ok {
			log.Fatal(fmt.Sprintf("Owner to tribe map has duplicate player id %v already mapped to %v", playerData.PlayerId, mappedTribe))
		}
		ownerTribeMap[playerData.PlayerId] = playerData.Tribe
	}

	return ownerTribeMap
}

func readAllActions(streamReader *io.SectionReader) map[int][]ActionCaptureCity {
	numActions := unsafeReadUint16(streamReader)

	turnCaptureMap := make(map[int][]ActionCaptureCity)

	turn := 1
	for i := 0; i < int(numActions); i++ {
		actionType := unsafeReadUint16(streamReader)

		if actionType == 1 {
			action := ActionBuild{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 2 {
			action := ActionAttack{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 3 {
			action := ActionRecover{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 4 {
			_ = readFixedList(streamReader, 9)
		} else if actionType == 5 {
			action := ActionTrain{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 6 {
			action := ActionMove{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 7 {
			action := ActionCaptureCity{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}

			_, ok := turnCaptureMap[turn]
			if !ok {
				turnCaptureMap[turn] = make([]ActionCaptureCity, 0)
			}
			turnCaptureMap[turn] = append(turnCaptureMap[turn], action)
		} else if actionType == 8 {
			action := ActionResearch{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 9 {
			action := ActionDestroyImprovement{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 11 {
			action := ActionCityReward{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 13 {
			action := ActionPromote{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 14 {
			action := ActionExamineRuins{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 15 {
			action := ActionEndTurn{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}

			if action.PlayerId == 255 {
				turn++
			}
		} else if actionType == 16 {
			action := ActionUpgrade{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 17 {
			_ = readFixedList(streamReader, 9)
		} else if actionType == 18 {
			_ = readFixedList(streamReader, 9)
		} else if actionType == 20 {
			_ = readFixedList(streamReader, 1)
		} else if actionType == 21 {
			action := ActionCityLevelUp{}
			if err := binary.Read(streamReader, binary.LittleEndian, &action); err != nil {
				log.Fatal("Failed to load action: ", err)
			}
		} else if actionType == 24 {
			_ = readFixedList(streamReader, 9)
		} else if actionType == 25 {
			_ = readFixedList(streamReader, 9)
		} else if actionType == 27 {
			_ = readFixedList(streamReader, 10)
		} else if actionType == 28 {
			_ = readFixedList(streamReader, 3)
		} else if actionType == 29 {
			_ = readFixedList(streamReader, 10)
		} else if actionType == 30 {
			_ = readFixedList(streamReader, 10)
		} else {
			log.Fatal("Unknown action type:", actionType)
		}
	}

	return turnCaptureMap
}

func ReadPolytopiaSaveFile(inputFilename string) (*PolytopiaSaveOutput, error) {
	decompressedReader, decompressedLength := buildReaderForDecompressedFile(inputFilename)
	streamReader := io.NewSectionReader(decompressedReader, int64(0), int64(decompressedLength))

	// Read initial map state
	initialMapHeaderOutput := DeserializeMapHeaderFromBytes(streamReader)

	initialTileData := make([][]TileData, initialMapHeaderOutput.MapHeight)
	for i := 0; i < initialMapHeaderOutput.MapHeight; i++ {
		initialTileData[i] = make([]TileData, initialMapHeaderOutput.MapWidth)
	}
	readTileData(streamReader, initialTileData, initialMapHeaderOutput.MapWidth, initialMapHeaderOutput.MapHeight)

	ownerTribeMap := buildOwnerTribeMap(readAllPlayerData(streamReader))

	_ = readFixedList(streamReader, 3)

	// Read current map state
	currentMapHeaderOutput := DeserializeMapHeaderFromBytes(streamReader)

	tileData := make([][]TileData, currentMapHeaderOutput.MapHeight)
	for i := 0; i < currentMapHeaderOutput.MapHeight; i++ {
		tileData[i] = make([]TileData, currentMapHeaderOutput.MapWidth)
	}
	readTileData(streamReader, tileData, currentMapHeaderOutput.MapWidth, currentMapHeaderOutput.MapHeight)

	playerData := readAllPlayerData(streamReader)
	ownerTribeMap = buildOwnerTribeMap(playerData)

	_ = readFixedList(streamReader, 2)

	turnCaptureMap := readAllActions(streamReader)

	output := &PolytopiaSaveOutput{
		MapHeight:       currentMapHeaderOutput.MapHeight,
		MapWidth:        currentMapHeaderOutput.MapWidth,
		PlayerData:      playerData,
		OwnerTribeMap:   ownerTribeMap,
		InitialTileData: initialTileData,
		TileData:        tileData,
		MaxTurn:         int(currentMapHeaderOutput.MapHeaderInput.CurrentTurn),
		TurnCaptureMap:  turnCaptureMap,
	}
	return output, nil
}
