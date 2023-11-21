<style lang="less">
.authModal {
    background: url('../../public/assets/img/login.jpg') center center;
    background-size: cover;
}
.layout{
    color: #fff;
    font-size: 15px;
    /* border: 1px solid #d7dde4; */
    position: relative;
	border-radius: 4px;
	overflow: hidden;
	width: 100%;
	text-align: center;
}
.layout-content-center{
    display: inline-block;
    
    margin-top: 10%;

	max-width: 500px;
}
.qrcode{
    display: inline-block;
	padding-top: 20px;

	/* img {
		width: 150px;
		height: 150px;
		background-color: #fff;
		padding: 6px;
		box-sizing: border-box;
	}; */
}
</style>
<template>
	<div>
		<div v-show="showModal" class="authModal">
			<Modal
				width="375"
				:closable="false"
				v-model="authModal"
				:mask="true"
				:mask-closable="false"
				class-name="vertical-center-modal">
				<h4 id="authTitle" style="text-align: center; margin-top: 5px;">{{$t("lang.authTitle")}}</h4>
				<div slot="footer" style="text-align: center; margin-bottom: 15px;">
					<div>
						<Row v-if="!authByPwd">
							<Col span="16">
								<Input v-model="authMail" type="email" :placeholder="$t('lang.inputEmail')"><span slot="prepend">{{$t("lang.email")}}</span></Input>
							</Col>
							<Col span="4">
								<Button id="authBtn" type="primary" @click="authCodeSend()" style="margin-left: 10px">{{$t("lang.sendCode")}}</Button>
							</Col>
						</Row>
						<Input v-if="!authByPwd" v-model="authCode" type="password" :placeholder="$t('lang.inputCode')" style="margin-top: 15px;"><span slot="prepend">{{$t("lang.code")}}</span></Input>
						<Input v-if="authByPwd" v-model="authMail" type="email" :placeholder="$t('lang.inputEmail')"><span slot="prepend">{{$t("lang.email")}}</span></Input>
						<Input v-if="authByPwd" v-model="authCode" type="password" :placeholder="$t('lang.inputCode')"  style="margin-top: 15px;"><span slot="prepend">{{$t("lang.password")}}</span></Input>
					</div>
					<div v-show="qrcodeImg !== ''" style="text-align: center; margin-top: 15px;">
						<img :src="qrcodeImg" alt="qrcode-img" />
						<!-- <div class="qrcode" ref="qrCodeUrl"></div> -->
					</div>
					<Button @click="runAuth()" type="primary" style="margin-top: 15px; margin-right: 10px; ">{{$t("lang.confirm")}}</Button>
					<Button @click="cancelAuth(false)" style="margin-top: 15px; margin-right: 10px; ">{{$t("lang.cancel")}}</Button>
					<div class="layout" style="margin-top: 20px">
						<a href="javascript:void(0)" @click="qrcodeImg = '';authByPwd = true;" style="margin-right: 10px; font-size: 16px;"><Icon type="md-code" style="margin-right: 5px;"/>Google</a>
						<a href="javascript:void(0)" @click="initAuth()" style="margin-right: 10px; font-size: 16px;"><Icon type="md-mail" style="margin-right: 5px;"/>{{$t("lang.email")}}</a>
						<a :disabled="qrcodeImg !== ''" href="javascript:void(0)" @click="initQRCodeImg()" style="margin-right: 10px; font-size: 16px;"><Icon type="md-qr-scanner" style="margin-right: 5px;"/>{{$t("lang.qrcode")}}</a>
					</div>
				</div>
			</Modal>
		</div>
	</div>
</template>
<script>
	import {webResponse} from '../toolchain/help.js';
    export default {
		inject: ["reload"],
        data () {
            return {
				user: null,
				userID: '',
				authMail: '',
				authCode: '',
				showModal: false,
				authModal: false,
				authByPwd: false,
				
				qrcodeAuth: false,
				justVerifyEmail: false,
				
				isPage: false,
				qrcodeImg: '',
				codetimer: null,
				redirectUrl: '/',
            }
		},
		props: {
			onAuthSuccessed: {
				type: Function,
				default: null
			},
			onUserOff: {
				type: Function,
				default: null
			},
		},
        mounted:function(){
			var cacheEmail = localStorage.getItem('authMail');
			if (cacheEmail !== undefined && cacheEmail !== null && typeof(cacheEmail) === 'string') {
				this.authMail = cacheEmail;
				console.log(this.authMail)
			}
			console.log(this.$route)
			
			if (window.location.href.indexOf('/auth') > -1) {
				this.isPage = true;
				if (this.$route.query.redirectUrl !== undefined) this.redirectUrl = this.$route.query.redirectUrl;
				this.runModal(this.$route.query.email)
			}
        },
        methods: {
			getUser() {
				return this.user;
			},
			runModal(email) {
				if (email !== undefined && email !== null && email !== '') this.justVerifyEmail = true;
				if (this.authMail === '') this.authMail = email;
				this.showModal = true;
				this.authModal = true;
			},
			cancelAuth(bSuccessed) {
				if (this.isPage === true) {
					if (bSuccessed === true) window.location.href = this.redirectUrl;
				} else {
					if (this.justVerifyEmail === false) localStorage.removeItem('isLogin');
					if (this.codetimer !== null) clearInterval(this.codetimer);
					this.justVerifyEmail = false;
					this.showModal = false;
					this.authModal = false;
					this.authCode = '';
				}
			},
			updateUser(response) {
				let authResult = false;
				if (response === null) {
					this.user = null;
					localStorage.removeItem('authMail');
					if (this.onUserOff !== null && this.onUserOff !== undefined) this.onUserOff(response);
				} else {
					authResult = true;
					this.user = response.data['Data'];
					localStorage.setItem('isLogin', 'true');
					localStorage.setItem('authMail', this.authMail);
					if (this.onAuthSuccessed !== null && this.onAuthSuccessed !== undefined) this.onAuthSuccessed(response);
				}
				this.cancelAuth(authResult);
			},
			// 这里的onUserOnline和onUserOff仅用于临时回调
			checkStatus(onAuthSuccessed, onUserOff, random) {
				let self = this;
				let getUrl = '/api/user/status';
				if (random !== undefined && random !== null && random !== '') getUrl = getUrl + '?random=' + random;
				this.httpGet(getUrl, function(response){
					webResponse(response, function(err){
						console.log(err)
						self.$Message.error(err)
						self.updateUser(null);
						if (onUserOff !== null && onUserOff !== undefined) onUserOff(response);
					}, function(response){
						self.updateUser(response);
						if (onAuthSuccessed !== null && onAuthSuccessed !== undefined) onAuthSuccessed(response);
					})
				})
			},
			logoutUser() {
				let self = this;
				this.httpGet('/api/user/logout', function(response){
					self.updateUser(null);
				}, function(response){
					self.updateUser(null);
				})
			},
			initAuth() {
				this.authByPwd = false;
				this.qrcodeImg = '';
			},
			authCodeSend() {
				if (this.checkAuthEmail() === false) return;

				let timeCount = 0;
				document.getElementById("authBtn").disabled = true;
				let oldName = document.getElementById("authBtn").innerHTML;
				let timer = setInterval(()=> {
					if (timeCount < 60000) {
						if (document.getElementById("authBtn") === undefined || document.getElementById("authBtn") === null) {
							clearInterval(timer);
							return;
						}
						document.getElementById("authBtn").innerHTML = (60000 - timeCount) / 1000 + 's ' + this.$t('lang.retry')
						timeCount = timeCount + 1000;
						return
					}
					clearInterval(timer);
					document.getElementById("authBtn").disabled = false;
					document.getElementById("authBtn").innerHTML = oldName;
				}, 1000)

				let self = this;
				let formdata = new FormData();
				formdata.append('kind', 'send');
				formdata.append('authID', this.authMail);
				this.httpPost('/api/user/auth', formdata, function(response){
					webResponse(response, function(err){
						self.$Message.error(err)
					})
				})
			},
			runAuth() {
				if (this.checkAuthEmail() === false) return;

				let self = this;
				let formdata = new FormData();
				let postUrl = '/api/user/auth';				
				formdata.append('kind', 'verify');
				formdata.append('authID', this.authMail);
				formdata.append('pwd', this.authCode);
				formdata.append('justVerify', this.justVerifyEmail);
				this.httpPost(postUrl, formdata, function(response){
					webResponse(response, function(err){self.$Message.error(err)}, self.authSuccessed)
				})
			},
			authSuccessed(response) {
				console.log(response, this.justVerifyEmail)
				this.$Message.success(this.$t('lang.authTitle') + this.$t('lang.successed'));
				if (this.justVerifyEmail === true) {
					if (this.codetimer !== null) clearInterval(this.codetimer);
					this.justVerifyEmail = false;
					this.showModal = false;
					this.authModal = false;
					if (this.onAuthSuccessed !== null && this.onAuthSuccessed !== undefined) this.onAuthSuccessed(response);
					return
				}
				this.updateUser(response);
			},
			checkAuthEmail() {
				let authMail = this.authMail;
				if (this.authMail === undefined || this.authMail === null) authMail = '';
				let isInvalidEmail = authMail !== '' && authMail.indexOf('@') === -1;
				if (authMail === '' || isInvalidEmail) {
					this.$Message.error(this.$t('lang.inputEmail'));
					return false
				}
				return true;
			},
			initQRCodeImg() {
				if (this.checkAuthEmail() === false) return;

				let self = this;
				let formdata = new FormData();
				formdata.append('kind', 'qrcode');
				formdata.append('authID', this.authMail);
				this.httpPost('/api/user/auth', formdata, function(response){
					webResponse(response, function(err){self.$Message.error(err)}, self.updateQRCodeImg)
				})
			},
			updateQRCodeImg(response) {
				this.qrcodeImg = response.data['Data'];
			},
			httpGet(url, onResponse, onPanic) {
				this.$axios.get(url)
				.then(function (response) {
					if (onResponse !== undefined && onResponse !== null) onResponse(response);
				})
				.catch(function (response) {
					if (onPanic !== undefined && onPanic !== null) onPanic(response);
				});
			},
			httpPost(url, formdata, onResponse, onPanic) {
				this.$axios.post(url, formdata)
				.then(function (response) {
					if (onResponse !== undefined && onResponse !== null) onResponse(response);
				})
				.catch(function (response) {
					if (onPanic !== undefined && onPanic !== null) onPanic(response);
				});
			}
        }
    }
    </script>