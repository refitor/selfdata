<style lang="less">
.nav-header-logo{
    font-size: 20px;
    text-align: right;
    margin-top: 15px;
}
.tr{text-align:right}
.mt10{
    margin-top: 10px;
}
.menu-item {
    margin-top: 10px;
    margin-bottom: 10px;
    font-size: 16px;
}
</style>
<template>
    <div>
        <Row style="height: 60px; background-color: #ebe9e7; vertical-align: middle;">
            <Col span="12">
                <div style="text-align: right;"> 
                    <div class="nav-header-logo">
                        <Icon :type="activeIcon" style="margin-right: 5px;"/>{{activeTile}}
                    </div>
                </div>
            </Col>
            <Col span="12">
                <div style="text-align: right; margin-top: 5px;"> 
                    <Button v-if="activePage === 'view'" @click="returnPage()" :icon="activePage === 'selfData' ? 'md-sync':'md-arrow-back'" style="margin-top: 10px; margin-right: 10px;"></Button>
                    <!-- <Button @click="syncData()" icon="md-lock" style="margin-top: 10px; margin-right: 10px;"></Button> -->
                    <Button @click="openOperate()" icon="md-open" style="margin-top: 10px; margin-right: 10px;"></Button>
                    <Button @click="showMenu = true" icon="ios-menu" style="margin-top: 10px; margin-right: 10px;"></Button>
                </div>
            </Col>
        </Row>
        <Drawer width="130" :closable="false" v-model="showMenu" placement="left">
            <p class="menu-item"><a href="javascript:void(0)" @click="reload()"><Icon type="md-unlock" style="margin-right: 5px;"/>{{$t('lang.decryption')}}</a></p>
            <Divider />
            <p class="menu-item"><a href="javascript:void(0)" @click="syncData()"><Icon type="md-lock" style="margin-right: 5px;"/>{{$t('lang.encryption')}}</a></p>
            <Divider />
            <p class="menu-item"><a href="javascript:void(0)" @click="openConf()"><Icon type="ios-settings" style="margin-right: 5px;"/>{{$t('lang.settings')}}</a></p>
            <Divider />
            <p class="menu-item"><a href="javascript:void(0)" @click="logout()"><Icon type="md-exit" style="margin-right: 5px;"/>{{$t('lang.logout')}}</a></p>
            <Divider />
        </Drawer>
        <div :style="{height: pageHeight() - 70 + 'px'}">
            <object v-if="activePage === 'page'" :data="pageUrl" width='100%' height='100%' style="padding-top: 10px; padding-left: 10px; padding-right: 10px;"/>
            <ConfPanel v-if="activePage === 'conf'" ref="confPanel" :formTitle="$t('lang.settings')" :updateFile="updateFile"></ConfPanel>
            <!-- <ConfPanel v-if="activePage === 'conf'" :params="pageParams" style="padding: 10px;"></ConfPanel> -->
            <ExplorerPanel v-show="activePage === 'selfData'" ref="selfdataPanel" :beforeOpen="beforeOpen" :from="'selfData'" :readonly="false" style="padding: 10px;"></ExplorerPanel>
            <EditorPanel v-if="activePage === 'view'" ref="viewPanel" :params="pageParams" style="padding: 10px;"></EditorPanel>
        </div>
        <Modal
            v-model="uploadModal"
            :footer-hide="hideFooter"
            class-name="vertical-center-modal">
            <p style="text-align: center;margin-bottom: 10px;">{{$t('lang.upload')}}{{$t('lang.file')}}</p>
            <Upload
                multiple
                type="drag"
                :on-success="uploadSuccessed"
                :action="'/api/file/upload?from=' + activePage + '&path=' + activeSDPath">
                <div style="padding: 20px 0">
                    <Icon type="ios-cloud-upload" size="52" style="color: #3399ff"></Icon>
                    <p>Click or drag files here to upload</p>
                </div>
            </Upload>
        </Modal>
        <Modal
            v-model="operateModal"
            :footer-hide="hideFooter"
            class-name="vertical-center-modal">
            <p style="text-align: center;margin: 5px;">{{$t('lang.resOperate')}}</p>
            <p style="text-align: left;margin: 10px 0;">{{$t('lang.path')}}: /{{activePage === 'view'? activeFilePath:activeSDPath}}</p>
            <Input v-if="canOperate('create-folder') || canOperate('create-file') || canOperate('rename')" v-model="resOperate" :placeholder="isEncryption() ? $t('lang.authIDPlaceholder') : $t('lang.operatePlaceholder')">
                <Select v-model="selectResKind" slot="append" style="width: 100px" @on-select="selectOperate" :placeholder="$t('lang.pleaseSelect')">
                    <Option v-show="canOperate('create-folder')" value="create-folder"><Icon type="md-add"/> {{$t('lang.createFolder')}}</Option>
                    <Option v-show="canOperate('create-file')" value="create-file"><Icon type="md-add"/> {{$t('lang.createFile')}}</Option>
                    <Option v-show="canOperate('rename')" value="rename"><Icon type="md-create"/> {{$t('lang.rename')}}</Option>
                </Select>
            </Input>
            <div style="text-align: center; margin-top: 5px;"> 
                <Button v-show="canOperate('update')" @click="updateFile()" icon="md-sync" type="info" style="margin-top: 10px; margin-right: 10px;">{{$t('lang.update')}}</Button>
                <Button v-show="canOperate('upload')" @click="uploadFile()" type="success" icon="md-cloud-upload" style="margin-top: 10px; margin-right: 10px;">{{$t('lang.upload')}}</Button>
                <Button v-show="canOperate('delete')" @click="operateRes('delete', activeSDPath, '')" type="error" icon="md-trash" style="margin-top: 10px; margin-right: 10px;">{{$t('lang.delete')}}</Button>
            </div>
        </Modal>
    </div>
</template>
<script>
import ConfPanel from './conf.vue';
import EditorPanel from './editor.vue';
import ExplorerPanel from './explorer.vue';
import {webResponse} from '../toolchain/help.js';
export default {
    inject: ["reload"],
    data(){
        return {
            operateModal: false,
            uploadModal: false,
            hideFooter: true,

            showMenu: false,
            activeTile: this.$t('lang.selfdata'),
            activePage: 'selfData',
            activeIcon: 'md-unlock',

            pageUrl: '',
            pageParams: {},
            resRootDir: 'encryption',
            activeSDPath: 'encryption',
            activeFilePath: '',

            // operate
            resOperate: '',
            activeOperate: '',
            selectResKind: '',

            // conf
            confData: {},
        }
    },
    mounted() {
        let self = this;
        if (this.$route.query.page === undefined) {
            this.httpGet('/api/user/status', function(response){
                webResponse(response, function(response){
                    self.$router.push({path: '/auth'})
                }, function(response){
                    self.init()
                })
            })
        } else {
            self.init()
        }
    },
    methods:{
        init() {
            var page = this.$route.query.page;
            var from = this.$route.query.from;
            var ftype = this.$route.query.ftype;
            var readonly = this.$route.query.readonly;
            var filePath = this.$route.query.filepath;
            if (page === 'view') {
                this.activeFilePath = filePath;
                window.location.hash = '#/';
                if (from !== undefined) this.pageParams['from'] = from;
                if (ftype !== undefined) this.pageParams['fType'] = ftype;
                if (readonly !== undefined) this.pageParams['readonly'] = readonly;
                if (filePath !== undefined) this.pageParams['filePath'] = filePath;
                this.openPage('view', this.$t('lang.fpreview'), 'ios-paper');
            } else if (page === 'conf') {
                this.openPage('conf', this.$t('lang.settings'), 'ios-settings');
            } else {
                this.initConf();
                this.openPage('selfData', this.$t('lang.selfdata'), 'md-unlock');
            }
        },
        initConf() {
            let self = this;
            let fileUrl = '/api/file/view?filepath=encryption/meta.json' + '&from=selfdata';
            this.httpGet(fileUrl, function(response) {
                if (response.data['Error'] !== '' && response.data['Error'] !== null && response.data['Error'] !== undefined) {
                    self.$Message.error(response.data['Error'])
                    self.close()
                } else {
                    self.confData = response.data;
                }
            }, null);
        },
        isEncryption() {
            let sysRight = this.confData["SysRight"] + '';
            if (sysRight.indexOf('strong') >= 0) {
                let workPath = this.activeSDPath;
                if (workPath === undefined || workPath === '') workPath = 'encryption';
                console.log('-----------', workPath)
                // console.log(this.$route.path, this.activeSDPath, workPath)
                return workPath === 'encryption' && this.activePage !== 'view'
            }
            return false
        },
        isRoot() {
            let workPath = this.activeSDPath;
            if (workPath === undefined) workPath = 'encryption';
            // console.log(this.$route.path, this.activeSDPath, workPath)
            return workPath.split('/').length === 2 && this.activePage !== 'view'
        },
        isFolder() {
            let workPath = this.activeSDPath;
            if (workPath === undefined) workPath = 'encryption';
            // console.log(this.$route.path, this.activeSDPath, workPath)
            return workPath.split('/').length > 2 && this.activePage !== 'view'
        },
        getState(attr) {
            if (attr === 'isRoot') return this.isRoot();
            else if (attr === 'isFolder') return this.isFolder();
        },
        openConf() {
            this.openPage('conf', this.$t('lang.settings'), 'ios-settings')
            // this.openPage('conf', this.$t('lang.settings'), 'ios-settings');
            // window.location.href = '/api/file/preview?page=view&readonly=false&filepath=encryption/meta.json&from=selfData'
        },
        openPage(name, title, icon, pageUrl) {
            if (pageUrl !== undefined) this.pageUrl = pageUrl;
            this.activeTile = title;
            this.activePage = name;
            this.activeIcon = icon;
            this.showMenu = false;
        },
        beforeOpen(resPath) {
            if (resPath === undefined) return true;
            if (this.isRoot()) this.operateRes('decrypt', resPath, '') 
            localStorage.setItem('redirect', resPath)
            this.activeSDPath = resPath;
            return true
        },
        returnPage() {
            window.location.hash = '/';
            if (this.activePage === 'view') {
                this.openPage('selfData', this.$t('lang.selfdata'), 'md-unlock');
                this.$refs.selfdataPanel.render();
                return
            }
            this.reload()
        },
        canOperate(kind) {
            let redirect = localStorage.getItem('redirect');
            if (redirect !== null && redirect !== undefined && redirect !== '') {
                this.activeSDPath = redirect;
                // localStorage.removeItem('redirect');
            }
            switch (kind) {
                case 'create-folder':
                    return this.isEncryption() || this.isRoot() || this.isFolder()
                case 'create-file':
                    return this.isFolder()
                case 'rename':
                    return this.activePage === 'view' || this.isFolder()
                case 'update':
                    return this.activePage === 'view'
                case 'upload':
                    return this.isFolder()
                case 'delete':
                   return this.activePage === 'view' || this.isFolder()
            }
            return false
        },
        openOperate() {
            this.operateModal = true;
        },
        uploadFile() {
            this.uploadModal = true;
        },
        uploadSuccessed(response, file, fileList) {
            this.reload()
        },
        selectOperate(op) {
            if (op.value === 'create-folder' || op.value === 'create-file' || op.value === 'rename') {
                if (this.resOperate === '') {
                    this.$Message.error(this.$t('lang.operateEditError'))
                    return
                }
            }
            this.operateRes(op.value, this.activeSDPath, this.resOperate)
        },
        syncData() {
            let self = this;
            self.$Modal.confirm({
                title: self.$t('lang.sdSync'),
                content: self.$t('lang.selfdataSync') + ', ' + self.$t('lang.syncConfirm') + '?',
                okText: self.$t('lang.ok'),
                cancelText: self.$t('lang.cancel'),
                onOk: () => { 
                    console.log('encrypt: ' + self.activeSDPath)
                    self.operateRes('encrypt', self.activeSDPath, '') 
                },
            })
        },
        updateFile(path, data) {
            if (path === undefined && data === undefined) {
                path = this.$refs.viewPanel.filePath;
                data = this.$refs.viewPanel.getContent();
            }
            this.operateRes('update', path, data)
        },
        // kind: create-folder, create-file, delete, update, decrypt
        operateRes(kind, path, data) {
            if (this.activePage === 'view') path = this.activeFilePath;

            let self = this;
            let formdata = new FormData();
            formdata.append('path', path)
            formdata.append('data', data)
            formdata.append('kind', kind)
            this.httpPost('/api/resource/operate', formdata, function(response){
                webResponse(response, function(err){
                    self.$Message.error(err)    
                }, function(response){
                    if (response.data['Data'] !== 'pass') {
                        self.$Message.success(self.$t('lang.resOperateOK'))
                        self.reload()
                    }
                })
            })
        },
        logout() {
            this.httpGet('/api/user/logout', function(response){
                localStorage.removeItem('isLogin')
                localStorage.removeItem('redirect')
                window.location.reload()
            }, function(response){
                localStorage.removeItem('isLogin')
                localStorage.removeItem('redirect')
                window.location.reload()
            })
        },
        pageHeight(){ //函数：获取尺寸
            //获取浏览器窗口高度
            var winHeight=0;
            if (window.innerHeight){
                winHeight = window.innerHeight;
            }
            else if ((document.body) && (document.body.clientHeight)){
                winHeight = document.body.clientHeight;
            }
            //通过深入Document内部对body进行检测，获取浏览器窗口高度
            if (document.documentElement && document.documentElement.clientHeight){
                winHeight = document.documentElement.clientHeight;
            }
            //DIV高度为浏览器窗口的高度
            //document.getElementById("test").style.height= winHeight +"px";
            //DIV高度为浏览器窗口高度的一半
            // document.getElementById("test").style.height= winHeight/2 +"px";
            return winHeight;
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
        ConfPanel,
        EditorPanel,
        ExplorerPanel,
    }
}
</script>
