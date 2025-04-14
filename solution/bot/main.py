### заранее прошу прощения за все что вы тут увидите, за использование п*тона и за НАСТОЛЬКО грязный код
### зато все работает

import logging
import asyncio
import aiohttp
from aiogram import Bot, Dispatcher, types, F
from aiogram.client.default import DefaultBotProperties
from aiogram.types import InlineKeyboardMarkup, InlineKeyboardButton, ReplyKeyboardMarkup, KeyboardButton
from aiogram.enums import ParseMode
from aiogram.filters import Command, or_f
from aiogram.utils.keyboard import InlineKeyboardBuilder, ReplyKeyboardBuilder
from aiogram.types import CallbackQuery, Message
import json
from dotenv import load_dotenv
import os

load_dotenv()
API_URL = os.environ.get("API_URL")
TOKEN = os.environ.get("TOKEN")

bot = Bot(token=TOKEN, default=DefaultBotProperties(parse_mode=ParseMode.HTML))
dp = Dispatcher()
logging.basicConfig(level=logging.INFO)

campaigns_PER_PAGE = 2

user_states = {}

async def fetch_campaigns(page, chat_id):
    global user_states
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{API_URL}/advertisers/{user_states[chat_id]["company_uuid"]}/campaigns?page={page}&size={campaigns_PER_PAGE}") as resp:
            return int(resp.headers.get("X-Total-Count")), (await resp.json())

async def fetch_campaign(campaign_uuid, chat_id):
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{API_URL}/advertisers/{user_states[chat_id]["company_uuid"]}/campaigns/{campaign_uuid}") as resp:
            return resp.status, (await resp.json())

async def fetch_delete_campaign(campaign_uuid, chat_id):
    async with aiohttp.ClientSession() as session:
        async with session.delete(f"{API_URL}/advertisers/{user_states[chat_id]["company_uuid"]}/campaigns/{campaign_uuid}") as resp:
            return resp.status

async def fetch_put_campaign(campaign_uuid, chat_id, body):
    async with aiohttp.ClientSession() as session:
        async with session.put(f"{API_URL}/advertisers/{user_states[chat_id]["company_uuid"]}/campaigns/{campaign_uuid}", json=body) as resp:
            return resp.status

async def get_advertiser(advertisers_uuid):
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{API_URL}/advertisers/{advertisers_uuid}") as resp:
            return resp.status, (await resp.json())

async def fetch_post_campaign(chat_id, body):
    async with aiohttp.ClientSession() as session:
        async with session.post(f"{API_URL}/advertisers/{user_states[chat_id]["company_uuid"]}/campaigns", json=body) as resp:
            return resp.status, (await resp.json())

async def get_advertiser_stats(chat_id):
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{API_URL}/stats/advertisers/{user_states[chat_id]["company_uuid"]}/campaigns") as resp:
            return resp.status, (await resp.json())

async def get_campaign_stats(campaign_uuid):
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{API_URL}/stats/campaigns/{campaign_uuid}") as resp:
            return resp.status, (await resp.json())

@dp.message(Command("start"))
async def send_welcome(message: Message):
    await message.answer("Привет!\n\nВведи uuid компании к которой ты принадлежишь, без этого никакие функции не будут доступны")
    user_states[message.chat.id] = {}
    user_states[message.chat.id]["stage"] = "select_company"

@dp.message(Command("help"))
async def send_welcome(message: Message):
    await message.answer("/list - получить объявления\n/create - создать новое объявление\n/stats - статистика по компании\n/cancel - отменить текущее действие\n/start - сбросить сессию")

@dp.message(or_f(Command("list"), F.text=="📋 Список объявлений"))
async def get_campaigns(message: Message):
    await build_campaigns_list(message.chat.id, 0)

async def build_campaigns_list(chat_id, page, msg_id=None):
    total, campaigns = await fetch_campaigns(page, chat_id)
    
    if not campaigns:
        await bot.send_message(chat_id, "у данной кампании нет объявлений\nсоздать новое - /create")
        return
    campaigns_reply = ""
    for campaign in campaigns:
        campaigns_reply += f"<b>{campaign["ad_title"]}</b>\n{campaign["ad_text"]}\nАктивна с {campaign["start_date"]} по {campaign["end_date"]} день\n/campaign_{campaign["campaign_id"].replace("-", "_")}\n\n"
    
    pagination_kb = InlineKeyboardBuilder()
    if page >= 1:
        pagination_kb.add(InlineKeyboardButton(text="⬅ Назад", callback_data=f"page_{page-1}"))
    if campaigns_PER_PAGE * (page+1) < total:
        pagination_kb.add(InlineKeyboardButton(text="Вперед ➡", callback_data=f"page_{page+1}"))
    if msg_id != None:
        await bot.edit_message_text(campaigns_reply, chat_id=chat_id, message_id=msg_id)
        await bot.edit_message_reply_markup(chat_id=chat_id, message_id=msg_id, reply_markup=pagination_kb.as_markup())
    else:
        await bot.send_message(chat_id, campaigns_reply, reply_markup=pagination_kb.as_markup())
@dp.message(F.text.startswith("/campaign_"))
async def show_campaign(message: Message):
    msg, kbd = await build_campaign_msg(message.text.replace("/campaign_", "").replace("_", "-"), message.chat.id)

    await bot.send_message(message.chat.id, msg, reply_markup=kbd)

async def build_campaign_msg(campaign_id, chat_id):
    status, campaign = await fetch_campaign(campaign_id, chat_id)
    if status != 200:
        return "Объявление не найдено", None
    kbd = InlineKeyboardBuilder()
    kbd.add(InlineKeyboardButton(text="✏️ Редактировать", callback_data=f"campaign_edit_{campaign["campaign_id"]}"))
    kbd.add(InlineKeyboardButton(text="📊 Статистика", callback_data=f"campaign_stats_{campaign["campaign_id"]}"))
    kbd.add(InlineKeyboardButton(text="🗑 Удалить", callback_data=f"campaign_delete_{campaign["campaign_id"]}"))
    msg = f'''Объявление
    <b>{campaign["ad_title"]}</b>
Текст объявления:
    {campaign["ad_text"]}
Активно с {campaign["start_date"]} по {campaign["end_date"]} день

Лимит показов: {campaign["impressions_limit"]}
Лимит кликов: {campaign["clicks_limit"]}
Цена за показ: {campaign["cost_per_impression"]}
Цена за клик: {campaign["cost_per_click"]}

Параметры таргетинга:
    Пол: {campaign["targeting"]["gender"]}
    Возраст с: {campaign["targeting"]["age_from"]} по {campaign["targeting"]["age_to"]} лет
    Локация: {campaign["targeting"]["location"]}'''
    return msg, kbd.as_markup()

@dp.message(or_f(Command("stats"), F.text=="📊 Статистика"))
async def campaign_stats(message: Message):
    status, stats = await get_advertiser_stats(message.chat.id)
    await message.answer(f"Статистика по компании за все время:\n\n{build_stats(stats)}")

def build_stats(stats):
    return f"Количество показов: {stats["impressions_count"]}\nКоличество кликов: {stats["clicks_count"]}\nКонверсия: {stats["conversion"]}%\nПотрачено на показы: {stats["spent_impressions"]}\nПотрачено на клики: {stats["spent_clicks"]}\nПотрачено всего: {stats["spent_total"]}"

@dp.callback_query(F.data.contains("page_"))
async def paginate_campaigns(callback_query: CallbackQuery):
    chat_id = callback_query.message.chat.id
    if chat_id not in user_states:
        return
    await build_campaigns_list(chat_id, int(callback_query.data.replace("page_", "")), callback_query.message.message_id)
    await callback_query.answer()

@dp.callback_query(F.data.startswith("campaign_delete_"))
async def delete_campaign(callback_query: CallbackQuery):
    campaign_uuid = callback_query.data.replace("campaign_delete_", "").replace("_", "-")
    await fetch_delete_campaign(campaign_uuid, callback_query.message.chat.id)
    await bot.send_message(callback_query.message.chat.id, "Объявление удалено")
    await callback_query.answer()

@dp.callback_query(F.data.startswith("campaign_edit_"))
async def edit_campaign(callback_query: CallbackQuery):
    campaign_id = callback_query.data.removeprefix("campaign_edit_")
    await bot.edit_message_text(chat_id=callback_query.message.chat.id, message_id=callback_query.message.message_id, text=callback_query.message.text + "\nВыберите что хотите отредактировать:")
    kbd = InlineKeyboardBuilder()
    kbd.add(InlineKeyboardButton(text="название", callback_data=f"edit_{campaign_id}_ad_title"))
    kbd.add(InlineKeyboardButton(text="описание", callback_data=f"edit_{campaign_id}_ad_text"))
    kbd.add(InlineKeyboardButton(text="$ за просмотр", callback_data=f"edit_{campaign_id}_cost_per_impression"))
    kbd.add(InlineKeyboardButton(text="$ за клик", callback_data=f"edit_{campaign_id}_cost_per_click"))
    kbd.add(InlineKeyboardButton(text="кол-во просмотров", callback_data=f"edit_{campaign_id}_impressions_limit"))
    kbd.add(InlineKeyboardButton(text="кол-во кликов", callback_data=f"edit_{campaign_id}_clicks_limit"))
    kbd.add(InlineKeyboardButton(text="возраст с", callback_data=f"edit_{campaign_id}_targeting_age_from"))
    kbd.add(InlineKeyboardButton(text="возраст по", callback_data=f"edit_{campaign_id}_targeting_age_to"))
    kbd.add(InlineKeyboardButton(text="гендер", callback_data=f"edit_{campaign_id}_targeting_gender"))
    kbd.add(InlineKeyboardButton(text="локация", callback_data=f"edit_{campaign_id}_targeting_location"))
    kbd.add(InlineKeyboardButton(text="отмена", callback_data=f"restore_campaign_{campaign_id}"))
    kbd.adjust(2,2,2,4,1)
    await bot.edit_message_reply_markup(chat_id=callback_query.message.chat.id, message_id=callback_query.message.message_id, reply_markup=kbd.as_markup())
    await callback_query.answer()

@dp.callback_query(F.data.startswith("campaign_stats_"))
async def stats_campaign(callback_query: CallbackQuery):
    campaign_id = callback_query.data.removeprefix("campaign_stats_")
    kbd = InlineKeyboardBuilder()
    kbd.add(InlineKeyboardButton(text="назад", callback_data=f"restore_campaign_{campaign_id}"))
    status, stats = await get_campaign_stats(campaign_id)
    await callback_query.message.edit_text(f"Статистика по объявлению за все время:\n\n{build_stats(stats)}", reply_markup=kbd.as_markup())
    await callback_query.answer()

@dp.callback_query(F.data.startswith("restore_campaign_"))
async def restore_campaign(callback_query: CallbackQuery):
    campaign_id = callback_query.data.removeprefix("restore_campaign_")
    msg, kbd = await build_campaign_msg(campaign_id, callback_query.message.chat.id)
    await callback_query.message.edit_text(msg, reply_markup=kbd)
    await callback_query.answer()

@dp.callback_query(F.data == "create_campaign_cancel")
async def create_campaign_cancel(callback_query: CallbackQuery):
    user_states[callback_query.message.chat.id]["new_campaign"] = {}
    await callback_query.message.answer("тогда не смею задерживать\n\nесли передумаешь - /create", message_effect_id="5046589136895476101")
    await callback_query.message.delete()
    await callback_query.answer()

@dp.callback_query(F.data == "create_campaign")
async def create_campaign_cancel(callback_query: CallbackQuery):
    await callback_query.message.delete()
    status, body = await fetch_post_campaign(callback_query.message.chat.id, user_states[callback_query.message.chat.id]["new_campaign"])
    user_states[callback_query.message.chat.id]["new_campaign"] = {}
    if status != 201:
        await callback_query.message.answer("эх, неудача... не переживай, в следующий раз обязательно получится!\n\nтык -> /create")
        return
    await callback_query.message.answer(f"успех! беги настраивать таргетинг в карточке объявления -> /campaign_{body["campaign_id"].replace("-", "_")}", message_effect_id="5046509860389126442")
    await callback_query.answer()

@dp.callback_query(F.data.startswith("edit_"))
async def edit_campaign_text(callback_query: CallbackQuery):
    edit_field = '_'.join(callback_query.data.split("_")[2:])
    await callback_query.message.answer(f"введи новое значение поля {edit_field}\n(чтобы отменить введи /cancel)")
    user_states[callback_query.message.chat.id]["stage"] = callback_query.data
    await callback_query.answer()

@dp.message(Command("cancel"))
async def cancel_command(message: Message):
    user_states[message.chat.id]["stage"] = None
    await message.answer("Текущее действие отменено")

@dp.message(or_f(Command("create"), F.text=="➕ Создать новое"))
async def cancel_command(message: Message):
    user_states[message.chat.id]["stage"] = "create_campaign_step_1"
    await message.answer("Отлично, давай начнем. Введи название компании\nОтменить можно в любой момент командой /cancel")

@dp.message()
async def handle_text(message: Message):
    if message.chat.id not in user_states.keys():
        await message.answer("не понял тебя, /help")
        return
    stage = user_states[message.chat.id]["stage"]
    if stage == None:
        await message.answer("не понял тебя, /help")
        return
    elif stage == "select_company":
        status, body = await get_advertiser(message.text)
        if status != 200:
            await message.answer("UUID компании не сходится с моей базой, попробуй еще")
            return
        user_states[message.chat.id]["company_uuid"] = message.text
        user_states[message.chat.id]["stage"] = None
        buttons = [
            [KeyboardButton(text="📋 Список объявлений")],
            [KeyboardButton(text="📊 Статистика")],
            [KeyboardButton(text="➕ Создать новое")],
        ]
        kbd = ReplyKeyboardMarkup(keyboard=buttons,resize_keyboard=True)
        await message.answer(f"запомнил, ты в компании {body["name"]}!\nознакомься со списком команд - /help", reply_markup=kbd)
    elif stage.startswith("edit_"):
        campaign_id = stage.split("_")[1]
        edit_field = '_'.join(stage.split("_")[2:])
        status, campaign = await fetch_campaign(campaign_id, message.chat.id)
        if status != 200:
            await message.answer("что то пошло не так...")
            return
        new_value = message.text
        if edit_field not in ["ad_title", "ad_text", "targeting_gender", "targeting_location"]:
            if edit_field in ["cost_per_impression", "cost_per_click"]:
                new_value = float(new_value)
            else:
                new_value = int(new_value)

        if edit_field.startswith("targeting_"):
            if "targeting" not in campaign:
                campaign["targeting"] = {}
            campaign["targeting"][edit_field.replace("targeting_", "")] = new_value
        else:
            campaign[edit_field] = new_value
        status = await fetch_put_campaign(campaign_id, message.chat.id, campaign)
        if status != 200:
            await message.answer(f"неверный формат ввода или поле сейчас изменить нельзя. попробуй ввести новое значение заново")
            return
        user_states[message.chat.id]["stage"] = None
        await message.answer("успешно изменено!", message_effect_id="5104841245755180586")
        msg, kbd = await build_campaign_msg(campaign_id, message.chat.id)
        await message.answer(msg, reply_markup=kbd)
    elif stage == "create_campaign_step_1":
        user_states[message.chat.id]["new_campaign"] = {}
        user_states[message.chat.id]["new_campaign"]["ad_title"] = message.text
        user_states[message.chat.id]["stage"] = "create_campaign_step_2"
        await message.answer(f"название объявления - \"{message.text}\", уже неплохо!\nтеперь напиши основной текст объявления")
    elif stage == "create_campaign_step_2":
        user_states[message.chat.id]["new_campaign"]["ad_text"] = message.text
        user_states[message.chat.id]["stage"] = "create_campaign_step_3"
        await message.answer(f"текст объявления:\n{message.text}\n\nтеперь задай желаемое количество просмотров и кликов, два числа через пробел")
    elif stage == "create_campaign_step_3":
        try:
            impressions_limit, clicks_limit = map(int, message.text.split())
        except:
            await message.answer("похоже что то не так с форматированием чисел, попробуй еще раз")
            return
        if clicks_limit > impressions_limit:
            await message.answer("желаемое количество кликов должно быть меньше количества просмотров!")
            return
        user_states[message.chat.id]["new_campaign"]["impressions_limit"] = impressions_limit
        user_states[message.chat.id]["new_campaign"]["clicks_limit"] = clicks_limit
        user_states[message.chat.id]["stage"] = "create_campaign_step_4"
        await message.answer(f"количество просмотров {impressions_limit}\nколичество кликов {clicks_limit}\n\nтеперь задай цену за просмотр и цену за клик, два числа (можно дробных) через пробел")
    elif stage == "create_campaign_step_4":
        try:
            per_impression, per_click = map(float, message.text.split())
        except:
            await message.answer("похоже что то не так с форматированием чисел, попробуй еще раз")
            return
        user_states[message.chat.id]["new_campaign"]["cost_per_impression"] = per_impression
        user_states[message.chat.id]["new_campaign"]["cost_per_click"] = per_click
        user_states[message.chat.id]["stage"] = "create_campaign_step_5"
        await message.answer(f"цена за просмотр {per_impression}\nцена за клик {per_click}\n\nу тебя отлично получается! последний шаг: теперь задай день начала и день окончания кампании, два числа через пробел. учти, что это единственный параметр который нельзя поменять в дальнейшем")
    elif stage == "create_campaign_step_5":
        try:
            start_date, end_date = map(int, message.text.split())
        except:
            await message.answer("похоже что то не так с форматированием чисел, попробуй еще раз")
            return
        user_states[message.chat.id]["new_campaign"]["start_date"] = start_date
        user_states[message.chat.id]["new_campaign"]["end_date"] = end_date
        user_states[message.chat.id]["stage"] = None
        await message.answer(f"день начала действия {start_date}\nдень конца действия {end_date}")
        kbd = InlineKeyboardBuilder()
        kbd.add(InlineKeyboardButton(text="отменить", callback_data="create_campaign_cancel"))
        kbd.add(InlineKeyboardButton(text="создать", callback_data="create_campaign"))
        await message.answer(f"так держать, осталось нажать на кнопку создать. удачи!\n\nНазвание: {user_states[message.chat.id]["new_campaign"]["ad_title"]}", reply_markup=kbd.as_markup())
    else:
        await message.answer("не понял тебя, /help")

async def main():
    await dp.start_polling(bot)

if __name__ == "__main__":
    asyncio.run(main())