<template>
    <div>
        <Modal
            :closable="false"
            v-model="rightOpenModal"
            :footer-hide="true"
            class-name="vertical-center-modal">
            <p style="text-align: center;margin-bottom: 10px;">{{$t('lang.open')}}</p>
            <Input v-model="bindEmail" type="email" readonly :placeholder="$t('lang.AuthIDPlaceholder')"><span slot="prepend">{{$t('lang.AuthID')}}</span></Input>
            <Input v-model="payAuthCode" type="text" :placeholder="$t('lang.AuthCodePlaceholder')" style="margin-top: 15px;"><span slot="prepend">{{$t('lang.AuthCode')}}</span></Input>
            <div style="text-align: center; margin-top: 15px;">
                <Button type="primary" @click="openNewTab('https://refitself.cn/#/price?product=selfdata&service=selfdata')" style="margin-right: 5px;">{{$t('lang.buy')}}</Button>
                <Button type="success" @click="verifyAuthCode()" style="margin-right: 5px;">{{$t('lang.confirm')}}</Button>
                <Button @click="rightOpenModal = false">{{$t('lang.cancel')}}</Button>
            </div>
        </Modal>
        <Modal
            width="800"
            :closable="false"
            v-model="formModal"
            :mask="true"
            :mask-closable="false"
            :ok-text="$t('lang.submit')"
            :cancel-text="$t('lang.cancel')"
            :footer-hide="false"
            @on-ok="submitForm"
            @on-cancel="close"
            class-name="vertical-center-modal">
            <p slot="header" style="color:#f60;text-align:center; margin-top: 10px;">
                <Icon type="ios-create-outline" />
                <span>{{formTitle}}</span>
            </p>
            <form-create ref="confForm" :rule="formData" v-model="fApi" :option="options" style="margin-left: 1px"/>
        </Modal>
        <AuthPanel ref="authPanel" :onAuthSuccessed="onVerifySuccessed"></AuthPanel>
         <a ref="mytarget" class="hidetarget" href="" target="_blank" rel="noopener noreferrer"></a>
    </div>
</template>

<script>
import AuthPanel from './auth.vue';
export default {
    inject: ["reload"],
    data(){
        return {
            fApi:{},
            formModal: false,
            options:{
                submitBtn: {show: false,}
            },
            formData: [],
            inputData: {},

            formJson: '',
            authIDEle: {},

            // buy
            rightOpenModal: false,
            payAuthCode: '',
            bindEmail: '',
        }
    },
    props: {
        formTitle: {
            type: String,
            default: ''
        }
    },
    mounted() {
        this.init()
        this.getConfigList('encryption/meta.json')
    }, 
    methods: {
        init() {
            if (this.$i18n.locale === 'en-US') this.formJson = JSON.stringify(require('../store/conf-en.json'))
            else if (this.$i18n.locale === 'zh-CN') this.formJson = JSON.stringify(require('../store/conf-zh.json'))
        },
        getConfigList(confPath) {
            let self = this;
            let fileUrl = '/api/file/view?filepath=' + confPath + '&from=selfdata';
            this.httpGet(fileUrl, function(response) {
                if (response.data['Error'] !== '' && response.data['Error'] !== null && response.data['Error'] !== undefined) {
                    self.$Message.error(response.data['Error'])
                    self.close()
                } else {
                    self.open(response.data)
                }
            }, null);
        },
        open(data) {
            let self = this;
            let formData = [];
            this.inputData = data;
            this.$i18n.locale = data['Language'];
            localStorage.setItem('lang', this.$i18n.locale)
            let tmpFormData = JSON.parse(this.formJson);
            tmpFormData.forEach(element => {
                for(let k in data) {
                    console.log('---------', k)
                    if (JSON.stringify(element).indexOf('verifyBtn') > -1) {
                        element.children[1].children[0].on = {
                            click:()=>{
                                self.clickBtn('verifyBtn')
                            }
                        }
                        self.authIDEle = element.children[0].children[0]
                    } else if (JSON.stringify(element).indexOf('openBtn') > -1) {
                        element.children[1].children[0].on = {
                            click:()=>{
                                self.clickBtn('openBtn')
                            }
                        }
                    }
                    if (JSON.stringify(element).indexOf(k) > -1) {
                        formData.push(element);
                        break
                    }
                    if (k === 'DataPublic' && JSON.stringify(element).indexOf('DataCert') > -1) {
                        formData.push(element);
                        break
                    }
                    if (k === 'NetPublic' && JSON.stringify(element).indexOf('NetCert') > -1) {
                        formData.push(element);
                        break
                    }
                }
            });
            this.bindEmail = data['AuthID']
            this.formData = formData;
            this.fApi.setValue(data)
            this.formModal = true;
        },
        close() {
            this.formData = [];
            this.formModal = false;
            window.location.reload()
        },
        clickBtn(kind) {
            if (kind === 'verifyBtn') {
                this.$refs.authPanel.runModal(this.inputData['AuthID'])
            } else if (kind === 'openBtn') {
                this.rightOpenModal = true;
                // this.openNewTab('https://refitself.cn/#/price?product=selfdata&service=selfdata')
            }
        },
        onVerifySuccessed(response) {
            this.authIDEle.props.readonly = false;
        },
        openNewTab(url) {
            let target = this.$refs.mytarget
            target.setAttribute('href', url)
            target.click()
        },
        verifyAuthCode() {
            this.inputData['PayBindEmail'] = this.bindEmail;
            this.inputData['PayAuthCode'] = this.payAuthCode;
            this.rightOpenModal = false;
            this.submitForm();
        },
        submitForm() {
            let self = this;
            for (let key in this.inputData) {
                let val = this.fApi.getValue(key)
                console.log(val)
                if (val !== undefined && val !== null) this.inputData[key] = val;
            }
            console.log(this.inputData)
            this.$parent.updateFile("encryption/meta.json", JSON.stringify(this.inputData));
            if (this.inputData['Language'] !== this.$i18n.locale) {
                this.$i18n.locale = this.inputData['Language'];
                localStorage.setItem('lang', this.$i18n.locale)
                this.reload();
            }
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
    },
    components: {
        AuthPanel
    }
}
</script>