package types

import (
	"github.com/wppzxc/wechat-tools/pkg/config"
	"time"
)

const (
	// EventGroupMsg 群消息
	EventGroupMsg = "EventGroupMsg"
	// EventFriendMsg 私聊消息
	EventFriendMsg = "EventFriendMsg"
	// EventSysMsg 系统消息
	EventSysMsg = "EventSysMsg"
	// EventGroupMemberAdd 群成员增加消息
	EventGroupMemberAdd = "EventGroupMemberAdd"
	// EventGroupMemberDecrease 群成员减少消息
	EventGroupMemberDecrease = "EventGroupMemberDecrease"
	// EventFriendVerify 好友请求
	EventFriendVerify = "EventFriendVerify"
)

// ---
// Type:
// 0: 系统消息
// 1: 文字消息
// 3: 图片消息
// 34: 语音消息
// 43: 视频消息
// 1/文本消息 3/图片消息 34/语音消息  42/名片消息  43/视频 47/动态表情 48/地理位置  49/分享链接  2001/红包  2002/小程序  2003/群邀请
const (
	SystemMsg        = 0
	TextMsg          = 1
	ImageMsg         = 3
	VoiceMsg         = 34
	CardMsg          = 42
	VideoMsg         = 43
	DynamicEmojiMsg  = 47
	PositionMsg      = 48
	ShareLinkMsg     = 49
	RedMoneyMsg      = 2001
	AppletsMsg       = 2002
	GroupInvite      = 2003
	FriendWelcomeMsg = 10000
)

const (
	DefaultRemoteHost     = "127.0.0.1"
	DefaultRemotePort     = "8073"
	DefaultRemoteEndPoint = "http://127.0.0.1:8073/httpAPI"

	//DefaultRemoteHost     = "152.136.224.208"
	//DefaultRemotePort     = "8073"
	//DefaultRemoteEndPoint = "http://152.136.224.208:8073/httpAPI"

	// DefaultRemoteHost     = "192.168.14.250"
	// DefaultRemotePort     = "8073"
	// DefaultRemoteEndPoint = "http://" + DefaultRemoteHost + ":" + DefaultRemotePort + "/httpAPI"
	DefaultTimeout = 5 * time.Second
)

// keaimao apiList
const (
	SendTextMsgApi          = "SendTextMsg"
	SendImageMsgApi         = "SendImageMsg"
	SendVideoMsgApi         = "SendVideoMsg"
	SendFileMsgApi          = "SendFileMsg"
	SendGroupMsgAndAtApi    = "SendGroupMsgAndAt"
	GetLoggedAccountListApi = "GetLoggedAccountList"
	GetFriendListApi        = "GetFriendList"
	GetGroupListApi         = "GetGroupList"
	GetGroupMemberList      = "GetGroupMemberList"
	RemoveGroupMember       = "RemoveGroupMember"
	AgreeFriendVerify       = "AgreeFriendVerify"
	AgreeGroupInvite        = "AgreeGroupInvite"
	InviteInGroup           = "InviteInGroup"
)

const (
	ManageMsgRemoveUser       = "踢"
	ManageMsgSetManager       = "设置管理员"
	ManageMsgRemoveManager    = "取消管理员"
	ManageMsgSetVip           = "设置白名单"
	ManageMsgSetUserInviteNum = "积分增加"
	ManageMsgHealthCheck      = "健康检查"
	ManageMsgInviteNumCheck   = "查询积分"
)

type RequestParam struct {
	Event         string `json:"event,omitempty"`
	Type          int    `json:"type,omitempty"`
	FromWxid      string `json:"from_wxid,omitempty"`
	FromName      string `json:"from_name,omitempty"`
	FinalFromWxid string `json:"final_from_wxid,omitempty"`
	FinalFromName string `json:"final_from_name,omitempty"`
	RobotWxid     string `json:"robot_wxid,omitempty"`
	Msg           string `json:"msg,omitempty"`
	JsonMsg       string `json:"json_msg,omitempty"`
}

type SendParam struct {
	Api        string `json:"api,omitempty"`
	Msg        string `json:"msg,omitempty"`
	Path       string `json:"path,omitempty"`
	ToWxid     string `json:"to_wxid,omitempty"`
	RobotWxid  string `json:"robot_wxid,omitempty"`
	IsRefresh  bool   `json:"is_refresh,omitempty"`
	MemberName string `json:"member_name,omitempty"`
	MemberWxid string `json:"member_wxid,omitempty"`
	CardData   string `json:"card_data,omitempty"`
	GroupWxid  string `json:"group_wxid,omitempty"`
}

type CommonResponseData struct {
	Code   int    `json:"Code"`
	Result string `json:"Result"`
}

type ResponseLocalUser struct {
	CommonResponseData
	Data string `json:"Data"`
}

type ResponseUserList struct {
	CommonResponseData
	Data []config.CommonUserInfo `json:"Data"`
}

type GroupUserDecJsonMsg struct {
	MemberWxid     string `json:"member_wxid"`
	MemberNickname string `json:"member_nickname"`
	GroupWxid      string `json:"group_wxid"`
	GroupName      string `json:"group_name"`
	Timestamp      int64  `json:"timestamp"`
}

type GroupUserAddJsonMsg struct {
	GroupWxid string  `json:"group_wxid"`
	GroupName string  `json:"group_name"`
	Guest     []Guest `json:"guest"`
	Inviter   Guest   `json:"inviter"`
}

type Guest struct {
	Wxid     string `json:"wxid"`
	Nickname string `json:"nickname"`
}

type TaoBaoApiResponse struct {
	TbkDgVegasTljCreateResponse TaoBaoTaoLiJinResponseData   `json:"tbk_dg_vegas_tlj_create_response"`
	TbkTpwdCreateResponse       TaoBaoTaoKouLingResponseData `json:"tbk_tpwd_create_response"`
}

type TaoBaoTaoLiJinResponseData struct {
	Result    TaoBaoResponseResult `json:"result"`
	RequestId string               `json:"request_id"`
}

type TaoBaoResponseResult struct {
	Model   TaoBaoResponseResultModel `json:"model"`
	Success bool                      `json:"success"`
	MsgCode string                    `json:"msg_code"`
	MsgInfo string                    `json:"msg_info"`
}

type TaoBaoResponseResultModel struct {
	RightsId  string `json:"rights_id"`
	SendUrl   string `json:"send_url"`
	VegasCode string `json:"vegas_code"`
}

type TaoBaoTaoKouLingResponseData struct {
	Data TaoBaoTaoKouLingResponseModelData `json:"data"`
}

type TaoBaoTaoKouLingResponseModelData struct {
	Model string `json:"model"`
}

type DaTaoKeResponse struct {
	Time int64
	Code int
	Msg  string
	Data []DaTaoKeItem
}

//id 				Number 	19259135 						商品id，在大淘客的商品id
//goodsId 			Number 	590858626868 					淘宝商品id
//ranking 			Number 	1 								榜单名次
//newRankingGoods 	Number 	1 								是否新上榜商品（12小时内入榜的商品） 0.否1.是
//dtitle 			String 	【李佳琦推荐】奢华芯肌素颜爆水霜 	短标题
//actualPrice 		Number 	39.9 							券后价
//commissionRate 	Number 	30 								佣金比例
//couponPrice 		Number 	300 							优惠券金额
//couponReceiveNum 	Number 	4000 							领券量
//couponTotalNum 	Number 	10000 							券总量
//monthSales 		Number 	8824 							月销量
//twoHoursSales 	Number 	1542 							2小时销量
//dailySales 		Number 	4545 							当天销量
//hotPush 			Number 	42 								热推值
//mainPic 			String 	“https://img.alicdn.com/i4/1687451966/O1CN01rTeKnv1QOTBnyOXDe\_!!1687451966.jpg“ 商品图
//title 			String 	“2019新款运动短裤女宽松防走光韩版外穿ins潮休闲学生bf夏季阔腿” 商品长标题
//desc 				String 	“多款可选！显瘦高腰韩版阔腿裤五分裤，不起球，不掉色。舒适面料，不挑身材，高腰设计” 商品描述
//originalPrice	 	Number 	29.9 							商品原价
//couponLink 		String 	“https://uland.taobao.com/quan/detail?sellerId=1687451966&activityId=ffef827d9a5747efbbe02a93c6d7ec13“ 优惠券链接
//couponStartTime 	String 	“2019-06-04 00:00:00” 			优惠券开始时间
//couponEndTime 	String 	“2019-06-06 23:59:59” 			优惠券结束时间
//commissionType 	Number 	3 								佣金类型
//createTime 		String 	“2019-06-03 17:55:18” 			创建时间
//activityType 		Number 	1 								活动类型
//picList 			Array 	“https://img.alicdn.com/imgextra/i2/1687451966/O1CN01WNuZcl1QOTCM9NsrO_!!1687451966.jpg,https://img.alicdn.com/imgextra/i4/1687451966/O1CN01h2ih4v1QOTCOxlZDj_!!1687451966.jpg“ 营销图
//guideName 		String 	易折网 							放单人名称
//shopType 			Number 	1 								店铺类型
//couponConditions 	Number 	29 								优惠券使用条件
//avgSales 			Number 	586 							日均销量（仅复购榜返回该字段）
//entryTime 		String 	“2019-06-06 10:59:59” 			入榜时间（仅复购榜返回该字段）
//sellerId 			String 	4014489195 						淘宝卖家id
//quanMLink 		Number 	10 								定金，若无定金，则显示0
//hzQuanOver 		Number 	100 							立减，若无立减金额，则显示0
//yunfeixian 		Number 	1 								0.不包运费险 1.包运费险
//estimateAmount 	Number 	25.2 							预估淘礼金
//freeshipRemoteDistrict Number 1 							偏远地区包邮，0.不包邮，1.包邮
//top 				Number 	1 								热词榜排名（适用于5.热词飙升榜6.热词排行榜）
//keyWord 			String 	螺蛳粉 							热搜词（适用于5.热词飙升榜6.热词排行榜）
//upVal 			Number 	1 								排名提升值（适用于5.热词飙升榜）
//hotVal 			Number 	123454 							排名热度值

type DaTaoKeItem struct {
	Id                     int64   `json:"id"`
	GoodsId                string  `json:"goodsId"`
	Ranking                int     `json:"ranking"`
	DTitle                 string  `json:"dtitle"`
	ActualPrice            float32 `json:"actualPrice"`
	CommissionRate         float32 `json:"commissionRate"`
	CouponPrice            float32 `json:"couponPrice"`
	CouponReceiveNum       int     `json:"couponReceiveNum"`
	CouponTotalNum         int     `json:"couponTotalNum"`
	MonthSales             int     `json:"monthSales"`
	TwoHoursSales          int     `json:"twoHoursSales"`
	DailySales             int     `json:"dailySales"`
	HotPush                int     `json:"hotPush"`
	MainPic                string  `json:"mainPic"`
	Title                  string  `json:"title"`
	Desc                   string  `json:"desc"`
	OriginalPrice          float32 `json:"originalPrice"`
	CouponLink             string  `json:"couponLink"`
	CouponStartTime        string  `json:"couponStartTime"`
	CouponEndTime          string  `json:"couponEndTime"`
	CommissionType         int     `json:"commissionType"`
	CreateTime             string  `json:"createTime"`
	ActivityType           int     `json:"activityType"`
	Imgs                   string  `json:"imgs"`
	GuideName              string  `json:"guideName"`
	ShopType               int     `json:"shopType"`
	CouponConditions       string  `json:"couponConditions"`
	NewRankingGoods        int     `json:"newRankingGoods"`
	SellerId               string  `json:"sellerId"`
	QuanMLink              int     `json:"quanMLink"`
	HzQuanOver             int     `json:"hzQuanOver"`
	Yunfeixian             int     `json:"yunfeixian"`
	EstimateAmount         float32 `json:"estimateAmount"`
	FreeshipRemoteDistrict int     `json:"freeshipRemoteDistrict"`
}
