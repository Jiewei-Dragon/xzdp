// let commonURL = "http://192.168.50.115:8081";
let commonURL = "/api";
// 设置后台服务地址
axios.defaults.baseURL = commonURL;
axios.defaults.timeout = 2000;
// 让 4xx/5xx 也进入成功回调，统一由业务字段 success 判断
axios.defaults.validateStatus = function (status) {
  return status >= 200 && status < 600;
}
// request拦截器，将用户token放入头中
axios.interceptors.request.use(
  config => {
    // 每次请求时动态获取 token
    const token = sessionStorage.getItem("token");
    if(token) {
      config.headers['Authorization'] = token;
    }
    return config
  },
  error => {
    console.log(error)
    return Promise.reject(error)
  }
)
function showError(msg){
  try {
    const now = Date.now()
    if (!window.__ERR_TOAST__) {
      window.__ERR_TOAST__ = { msg: '', ts: 0 }
    }
    // 同一文案在1.5秒内不重复弹
    if (window.__ERR_TOAST__.msg === msg && (now - window.__ERR_TOAST__.ts) < 1500) {
      return
    }
    window.__ERR_TOAST__.msg = msg
    window.__ERR_TOAST__.ts = now
  } catch (e) {}
  try{
    if (window.ELEMENT && window.ELEMENT.Message) {
      window.ELEMENT.Message.error(msg)
      return
    }
    if (window.Element && window.Element.Message) {
      window.Element.Message.error(msg)
      return
    }
  }catch(e){}
  alert(msg)
}

axios.interceptors.response.use(function (response) {
  // 判断执行结果
  if (!response.data?.success) {
    const d = response?.data;
    const msg = (typeof d === 'string' ? d : (d?.errorMsg || d?.message)) || '请求失败';
    return Promise.reject(msg)
  }
  return response.data;
}, function (error) {
  // 一般是服务端异常或者网络异常
  console.log(error)
  if (error?.response?.status == 401) {
    // 未登录，跳转
    setTimeout(() => {
      location.href = "/login.html"
    }, 200);
    return Promise.reject("请先登录");
  }
  const d = error?.response?.data;
  const msg = (typeof d === 'string' ? d : (d?.errorMsg || d?.message)) || error?.message || "服务器异常";
  return Promise.reject(msg);
});

// 如需兜底未捕获Promise错误，可按需开启：
// window.addEventListener('unhandledrejection', (event) => {
//   const r = event.reason
//   const msg = (typeof r === 'string' ? r : (r?.message || r?.errorMsg || r?.toString())) || '服务器异常'
//   showError(msg)
// })
axios.defaults.paramsSerializer = function(params) {
  let p = "";
  Object.keys(params).forEach(k => {
    if(params[k]){
      p = p + "&" + k + "=" + params[k]
    }
  })
  return p;
}
const util = {
  commonURL,
  getUrlParam(name) {
    let reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)", "i");
    let r = window.location.search.substr(1).match(reg);
    if (r != null) {
      return decodeURI(r[2]);
    }
    return "";
  },
  formatPrice(val) {
    if (typeof val === 'string') {
      if (isNaN(val)) {
        return null;
      }
      // 价格转为整数
      const index = val.lastIndexOf(".");
      let p = "";
      if (index < 0) {
        // 无小数
        p = val + "00";
      } else if (index === p.length - 2) {
        // 1位小数
        p = val.replace("\.", "") + "0";
      } else {
        // 2位小数
        p = val.replace("\.", "")
      }
      return parseInt(p);
    } else if (typeof val === 'number') {
      if (!val) {
        return null;
      }
      const s = val + '';
      if (s.length === 0) {
        return "0.00";
      }
      if (s.length === 1) {
        return "0.0" + val;
      }
      if (s.length === 2) {
        return "0." + val;
      }
      const i = s.indexOf(".");
      if (i < 0) {
        return s.substring(0, s.length - 2) + "." + s.substring(s.length - 2)
      }
      const num = s.substring(0, i) + s.substring(i + 1);
      if (i === 1) {
        // 1位整数
        return "0.0" + num;
      }
      if (i === 2) {
        return "0." + num;
      }
      if (i > 2) {
        return num.substring(0, i - 2) + "." + num.substring(i - 2)
      }
    }
  }
}
