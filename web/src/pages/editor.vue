<style lang="less">
.explorer{
    background-color: #ebe9e7;
}
</style>
<template>
  <div>
    <!-- <span style="top: 20px; position: absolute; color: #fc3b00;">{{$t('lang.editorNotify')}}</span> -->
    <Editor v-show="showEditor === true" ref="fileEditor"
      :content="content"
      :fontSize="14"
      :height="pageHeight() - 91 + 'px'"
      :lang="lang"
      theme="chrome"
      :width="pageWidth()"
      :options="editorOptions"
      :readonly="readonly === 'true'"
      @init="editorInit"
      @onChange="editorChange"
      @onInput="editorInput"
      @onFocus="editorFocus"
      @onBlur="editorBlur"
      @onPaste="editorPaste"
    ></Editor>
    <object class="explorer" v-if="showEditor === false" :data="explorerUrl" :width="pageWidth()" :height="pageHeight() - 70 + 'px'" />
  </div>
</template>

<script>
import Editor from 'vue2x-ace-editor';
import {webResponse} from '../toolchain/help.js';
export default {
  name: "editor",
  data() {
        return {
            lang: 'text',
            content: '',
            showEditor: false,
            explorerUrl: '',

            fType: '',
            from: '',
            readonly: '',
            filePath: '',

            editorOptions: {  // 设置代码编辑器的样式
                enableBasicAutocompletion: true,
                enableSnippets: true,
                enableLiveAutocompletion: true,
                tabSize: 2,
                fontSize: 16,
                showPrintMargin: false   //去除编辑器里的竖线
            }
        };
    },
    props: {
        params: {
            type: Object,
            default: {},
        }
    },
    mounted() {
        if (this.params !== undefined && this.params !== null) {
            if (this.params['from'] !== undefined) this.from = this.params['from'];
            if (this.params['fType'] !== undefined) this.fType = this.params['fType'];
            if (this.params['filePath'] !== undefined) this.filePath = this.params['filePath'];
            if (this.params['readonly'] !== undefined) this.readonly = this.params['readonly'];
        }

        // InitFooter();   
        this.showEditor = this.fType.indexOf('text/plain') !== -1;     
        if (this.filePath !== '' && this.showEditor === true) {
            var msgobj = this.$Message;       
            let updateContent = this.updateContent;
            let getUrl = "/api/file/view?readonly=" + this.readonly + "&filepath=" + this.filePath + "&ftype=" + this.fType + '&from=' + this.from;
            this.httpGet(getUrl, function(response){
                updateContent(response.data)
            })
        } 
        if (this.filePath !== '' && this.showEditor === false) {
            let explorerUrl = '/api/file/view?readonly=true&filepath=' + this.filePath + "&ftype=" + this.fType + '&from=' + this.from;
            this.explorerUrl = explorerUrl;
        }
    },
    methods: {
        isImg(){
            let isVideo = this.filePath.indexOf('.mp4') > 0 || name.indexOf('.mov') > 0;
            let isAudio = this.filePath.indexOf('.mp3') > 0 || name.indexOf('.wav') > 0;
            let isImg = this.filePath.indexOf('.png') > 0 || this.filePath.indexOf('.jpg') > 0 || this.filePath.indexOf('.jpeg') > 0 || this.filePath.indexOf('.gif') > 0 || this.filePath.indexOf('.bmp') > 0;
            return isImg || isVideo || isAudio;
        },
        getFileType(filename){
            if(!filename || typeof filename != 'string'){
                return ''
            };
            let a = filename.split('').reverse().join('');
            let b = a.substring(0,a.search(/\./)).split('').reverse().join('');
            return b
        },
        pageWidth(){ //函数：获取尺寸
            //获取浏览器窗口高度
            var winWidth=0;
            if (window.innerWidth){
                winWidth = window.innerWidth;
            }
            else if ((document.body) && (document.body.clientWidth)){
                winWidth = document.body.clientWidth;
            }
            //通过深入Document内部对body进行检测，获取浏览器窗口高度
            if (document.documentElement && document.documentElement.clientWidth){
                winWidth = document.documentElement.clientWidth;
            }
            //DIV高度为浏览器窗口的高度
            //document.getElementById("test").style.height= winHeight +"px";
            //DIV高度为浏览器窗口高度的一半
            // document.getElementById("test").style.height= winHeight/2 +"px";
            return winWidth;
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
        updateContent(data) {
            if (typeof(data) != 'string') {
                data = JSON.stringify(data, null, 2)
            }
            this.content = data;
        },
        getContent() {
            return this.$refs.fileEditor.getValue();
        },
        editorInit() {
            require("brace/ext/language_tools");
            require(`brace/mode/text`);
            require(`brace/snippets/text`);
            require(`brace/theme/chrome`);
            
            require('brace/mode/json')    //language
            require('brace/mode/less')
            require('brace/mode/yaml')
            require('brace/snippets/json') //snippet
        },
        editorChange(editor) {
            // console.log("changed", editor.getValue());
        },
        editorInput(editor) {
            // console.log("input", editor.getValue());
        },
        editorFocus(editor) {
            // console.log("focus", editor);
        },
        editorBlur(editor) {
            // console.log("blur", editor);
        },
        editorPaste(editor) {
            // console.log("pase", editor);
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
    components:{
        Editor
    }
};
</script>