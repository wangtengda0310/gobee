package wow

import (
	"bytes"
	"encoding/binary"
)

// AuthResult
const (
	WOW_SUCCESS                 = 0x00
	WOW_FAIL_UNKNOWN0           = 0x01 ///< ? Unable to connect
	WOW_FAIL_UNKNOWN1           = 0x02 ///< ? Unable to connect
	WOW_FAIL_BANNED             = 0x03 ///< This <game> account has been closed and is no longer available for use. Please go to <site>/banned.html for further information.
	WOW_FAIL_UNKNOWN_ACCOUNT    = 0x04 ///< The information you have entered is not valid. Please check the spelling of the account name and password. If you need help in retrieving a lost or stolen password, see <site> for more information
	WOW_FAIL_INCORRECT_PASSWORD = 0x05 ///< The information you have entered is not valid. Please check the spelling of the account name and password. If you need help in retrieving a lost or stolen password, see <site> for more information
	// client reject next login attempts after this error, so in code used WOW_FAIL_UNKNOWN_ACCOUNT for both cases
	WOW_FAIL_ALREADY_ONLINE  = 0x06 ///< This account is already logged into <game>. Please check the spelling and try again.
	WOW_FAIL_NO_TIME         = 0x07 ///< You have used up your prepaid time for this account. Please purchase more to continue playing
	WOW_FAIL_DB_BUSY         = 0x08 ///< Could not log in to <game> at this time. Please try again later.
	WOW_FAIL_VERSION_INVALID = 0x09 ///< Unable to validate game version. This may be caused by file corruption or interference of another program. Please visit <site> for more information and possible solutions to this issue.
	WOW_FAIL_VERSION_UPDATE  = 0x0A ///< Downloading
	WOW_FAIL_INVALID_SERVER  = 0x0B ///< Unable to connect
	WOW_FAIL_SUSPENDED       = 0x0C ///< This <game> account has been temporarily suspended. Please go to <site>/banned.html for further information
	WOW_FAIL_FAIL_NOACCESS   = 0x0D ///< Unable to connect
	WOW_SUCCESS_SURVEY       = 0x0E ///< Connected.
	WOW_FAIL_PARENTCONTROL   = 0x0F ///< Access to this account has been blocked by parental controls. Your settings may be changed in your account preferences at <site>
	WOW_FAIL_LOCKED_ENFORCED = 0x10 ///< You have applied a lock to your account. You can change your locked status by calling your account lock phone number.
	WOW_FAIL_TRIAL_ENDED     = 0x11 ///< Your trial subscription has expired. Please visit <site> to upgrade your account.
	WOW_FAIL_USE_BATTLENET   = 0x12 ///< WOW_FAIL_OTHER This account is now attached to a Battle.net account. Please login with your Battle.net account email address and password.
	// WOW_FAIL_OVERMIND_CONVERTED
	// WOW_FAIL_ANTI_INDULGENCE
	// WOW_FAIL_EXPIRED
	// WOW_FAIL_NO_GAME_ACCOUNT
	// WOW_FAIL_BILLING_LOCK
	// WOW_FAIL_IGR_WITHOUT_BNET
	// WOW_FAIL_AA_LOCK
	// WOW_FAIL_UNLOCKABLE_LOCK
	// WOW_FAIL_MUST_USE_BNET
	// WOW_FAIL_OTHER
)

func responseChallenge() {
	// 1 ip是否被封：AuthSocket.cpp:392-401 WOW_FAIL_FAIL_NOACCESS
	// 2 ip是否绑定：AuthSocket.cpp:441-428 WOW_FAIL_SUSPENDED
	// 3 账号被封：AuthSocket.cpp:432-448 WOW_FAIL_BANNED，WOW_FAIL_SUSPENDED
	// 4 account v（verify） s（salt） from db
	// 5 if (databaseV.size() != s_BYTE_SIZE * 2 || databaseS.size() != s_BYTE_SIZE * 2)
	//                        _SetVSFields(rI);
	// 6 生成152位随机数 b，gmod（部超过32位） = g ^ b % N,计算服务端公钥B = ((v * 3) + gmod) % N
}

var VersionChallenge = [16]byte{0xBA, 0xA3, 0x1E, 0x99, 0xA0, 0x0B, 0x21, 0x57, 0xFC, 0x37, 0x3F, 0xB3, 0x69, 0xCD, 0xD2, 0xF1}

type LoginChallengeResponse struct {
	Cmd              uint8
	Error            byte
	FailEnum         byte // 0x0d
	B                [32]byte
	GLen             byte
	G                byte
	NLen             byte
	N                [32]byte
	S                [32]byte
	VersionChallenge [16]byte
	SecurityFlags    byte // token const '4' add from 2.4.3
}

func (resp *LoginChallengeResponse) Marshal(data []byte) error {
	data[0] = resp.Cmd
	data[1] = resp.Error
	data[2] = resp.FailEnum
	copy(data[3:35], resp.B[:])
	data[35] = resp.GLen
	data[36] = resp.G
	data[37] = resp.NLen
	copy(data[38:70], resp.N[:])
	copy(data[70:102], resp.S[:])
	copy(data[102:118], resp.VersionChallenge[:])
	data[118] = resp.SecurityFlags
	return nil
}
func (resp *LoginChallengeResponse) UnMarshal(data []byte) error {
	resp.Cmd = data[0]
	resp.Error = data[1]
	resp.FailEnum = data[2]
	copy(resp.B[:], data[3:35])
	resp.GLen = data[35]
	resp.G = data[36]
	resp.NLen = data[37]
	copy(resp.N[:], data[38:70])
	copy(resp.S[:], data[70:102])
	copy(resp.VersionChallenge[:], data[102:118])
	resp.SecurityFlags = data[118]
	var r error = nil
	return r

}

type LoginChallengeResponse_2_4_3 struct {
	LoginChallengeResponse
	PinAboutData1 [32]byte
	PinAboutData2 [32]byte
	PinAboutData3 [64]byte
	UnkAboutData1 [8]byte
	UnkAboutData2 [8]byte
	UnkAboutData3 [8]byte
	UnkAboutData4 [8]byte
	UnkAboutData5 [64]byte
	Authenticator [0]byte
}

type LoginChallengeMsg struct {
	Cmd          uint8
	Error        uint8
	Size         uint16
	GameName     [4]byte
	Version1     uint8
	Version2     uint8
	Version3     uint8
	Build        uint16
	Platform     [4]uint8
	Os           [4]uint8
	Country      [4]uint8
	TimeZoneBias uint32
	Ip           [4]uint8
	ILen         uint8
	I            string
}

func (m *LoginChallengeMsg) Marshal(data []byte) error {
	data[0] = m.Cmd
	data[1] = m.Error
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, m.Size)
	copy(data[2:4], buf.Bytes())
	copy(data[4:8], m.GameName[:])
	data[8] = m.Version1
	data[9] = m.Version2
	data[10] = m.Version3
	var writer bytes.Buffer
	_ = binary.Write(&writer, binary.LittleEndian, m.Build)
	copy(data[11:13], writer.Bytes())
	data[13] = m.Platform[3]
	data[14] = m.Platform[2]
	data[15] = m.Platform[1]
	data[16] = m.Platform[0]
	data[17] = m.Os[3]
	data[18] = m.Os[2]
	data[19] = m.Os[1]
	data[20] = m.Os[0]
	data[21] = m.Country[3]
	data[22] = m.Country[2]
	data[23] = m.Country[1]
	data[24] = m.Country[0]
	var reader bytes.Buffer
	_ = binary.Write(&reader, binary.LittleEndian, m.TimeZoneBias)
	copy(data[25:29], reader.Bytes())
	copy(data[29:33], m.Ip[:])
	data[33] = m.ILen
	copy(data[34:34+m.ILen], []byte(m.I))
	var r error = nil
	return r
}

func (m *LoginChallengeMsg) UnMarshal(data []byte) error {
	m.Cmd = data[0]
	m.Error = data[1]
	reader := bytes.NewReader(data[2:4])
	_ = binary.Read(reader, binary.LittleEndian, &m.Size)
	copy(m.GameName[:], data[4:8])
	m.Version1 = data[8]
	m.Version2 = data[9]
	m.Version3 = data[10]
	reader = bytes.NewReader(data[11:13])
	_ = binary.Read(reader, binary.LittleEndian, &m.Build)
	m.Platform = [4]uint8{data[16], data[15], data[14], data[13]}
	m.Os = [4]uint8{data[20], data[19], data[18], data[17]}
	m.Country = [4]uint8{data[24], data[23], data[22], data[21]}
	reader = bytes.NewReader(data[25:29])
	_ = binary.Read(reader, binary.LittleEndian, &m.TimeZoneBias)
	copy(m.Ip[:], data[29:33])
	m.ILen = data[33]
	m.I = string(data[34 : 34+m.ILen])
	var r error = nil
	return r
}

type LoginProofRequest struct {
	Cmd           uint8
	A             [32]byte
	M1            [20]byte
	CRC1          [20]byte
	NumberOfKeys  uint8
	SecurityFlags uint8
}

func (r *LoginProofRequest) Marshal(data []byte) error {
	data[0] = r.Cmd
	copy(data[1:33], r.A[:])
	copy(data[33:53], r.M1[:])
	copy(data[53:73], r.CRC1[:])
	data[73] = r.NumberOfKeys
	data[74] = r.SecurityFlags

	return nil

}

func (r *LoginProofRequest) UnMarshal(data []byte) error {
	r.Cmd = data[0]
	copy(r.A[:], data[1:33])
	copy(r.M1[:], data[33:53])
	copy(r.CRC1[:], data[53:73])
	r.NumberOfKeys = data[73]
	r.SecurityFlags = data[74]

	return nil
}

type LoginProofResponse struct {
	Cmd          uint8
	Error        uint8
	M2           [20]byte
	accountFlags uint32
	surveyId     uint32
	unkFlags     uint16
}

func (r *LoginProofResponse) Marshal(data []byte) error {
	data[0] = r.Cmd
	data[1] = r.Error
	copy(data[2:22], r.M2[:])
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, r.accountFlags)
	copy(data[22:23], buf.Bytes())
	buf.Reset()
	_ = binary.Write(&buf, binary.LittleEndian, r.surveyId)
	copy(data[23:24], buf.Bytes())
	buf.Reset()
	_ = binary.Write(&buf, binary.LittleEndian, r.unkFlags)
	copy(data[24:25], buf.Bytes())
	return nil
}

func (r *LoginProofResponse) UnMarshal(data []byte) error {
	r.Cmd = data[0]
	r.Error = data[1]
	copy(r.M2[:], data[2:22])
	reader := bytes.NewReader(data[22:23])
	_ = binary.Read(reader, binary.LittleEndian, &r.accountFlags)
	reader = bytes.NewReader(data[23:24])
	_ = binary.Read(reader, binary.LittleEndian, &r.surveyId)
	reader = bytes.NewReader(data[24:25])
	_ = binary.Read(reader, binary.LittleEndian, &r.unkFlags)
	return nil
}
